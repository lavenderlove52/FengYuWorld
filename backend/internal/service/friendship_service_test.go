package service_test

import (
	"testing"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 好友申请的冲突分支依赖真实唯一约束（拒绝后重新申请要复用同一行），
// 所以这里连本地 PostgreSQL，每个用例在事务里跑并回滚。

func friendshipDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "postgres://myblog:myblog@localhost:5432/myblog?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skipf("skip: 连不上测试库 %v", err)
	}
	return db
}

func seedFriendUser(t *testing.T, tx *gorm.DB, name string) *domain.User {
	t.Helper()
	suffix := time.Now().UnixNano()
	user := &domain.User{
		Username:     name + "-" + itoaFriend(suffix),
		Email:        name + "-" + itoaFriend(suffix) + "@test.local",
		PasswordHash: "x",
		Nickname:     name,
	}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}

func itoaFriend(v int64) string {
	if v == 0 {
		return "0"
	}
	digits := make([]byte, 0, 20)
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	return string(digits)
}

func TestFriendship_SendRequestBranches(t *testing.T) {
	db := friendshipDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin: %v", tx.Error)
	}
	defer tx.Rollback()

	svc := service.NewFriendshipService(repository.NewFriendshipRepository(tx), repository.NewUserRepository(tx))
	alice := seedFriendUser(t, tx, "alice")
	bob := seedFriendUser(t, tx, "bob")

	req := dto.FriendRequestCreate{FriendUsername: bob.Username}

	// 1. 首次申请成功
	if _, err := svc.SendRequest(alice.ID, req); err != nil {
		t.Fatalf("首次申请应成功: %v", err)
	}

	// 2. 同一个人再发：应是「已发送过」
	if _, err := svc.SendRequest(alice.ID, req); err != service.ErrRequestPending {
		t.Fatalf("重复申请应返回 ErrRequestPending，得到 %v", err)
	}

	// 3. 对方反过来申请：应是「对方已发给我」
	reverse := dto.FriendRequestCreate{FriendUsername: alice.Username}
	if _, err := svc.SendRequest(bob.ID, reverse); err != service.ErrRequestIncoming {
		t.Fatalf("反向申请应返回 ErrRequestIncoming，得到 %v", err)
	}

	// 4. 接受后再申请：应是「已经是好友」
	received, err := svc.ListReceived(bob.ID)
	if err != nil || len(received) != 1 {
		t.Fatalf("bob 应收到 1 条申请，得到 %d（err=%v）", len(received), err)
	}
	if err := svc.Accept(bob.ID, received[0].ID); err != nil {
		t.Fatalf("接受申请失败: %v", err)
	}
	if _, err := svc.SendRequest(alice.ID, req); err != service.ErrAlreadyFriends {
		t.Fatalf("已是好友应返回 ErrAlreadyFriends，得到 %v", err)
	}

	// 5. 删好友 → 被拒绝 → 重新申请，应复用原行成功（原来会撞唯一键报系统错误）
	if err := svc.DeleteFriend(alice.ID, bob.ID); err != nil {
		t.Fatalf("删除好友失败: %v", err)
	}
	if _, err := svc.SendRequest(bob.ID, reverse); err != nil {
		t.Fatalf("删除后重新申请应成功: %v", err)
	}
	bobReceived, _ := svc.ListReceived(alice.ID)
	if len(bobReceived) != 1 {
		t.Fatalf("alice 应收到 1 条申请，得到 %d", len(bobReceived))
	}
	if err := svc.Reject(alice.ID, bobReceived[0].ID); err != nil {
		t.Fatalf("拒绝申请失败: %v", err)
	}
	item, err := svc.SendRequest(bob.ID, reverse)
	if err != nil {
		t.Fatalf("被拒绝后重新申请应成功（复用原行），得到 %v", err)
	}
	if item == nil {
		t.Fatal("重新申请应返回申请记录")
	}
	pending, _ := svc.ListReceived(alice.ID)
	if len(pending) != 1 {
		t.Fatalf("重新申请后 alice 应收到 1 条待处理申请，得到 %d", len(pending))
	}
	// 复用原行：记录还是一条，不应该多出一条
	var count int64
	if err := tx.Model(&domain.Friendship{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			alice.ID, bob.ID, bob.ID, alice.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("同一对用户应只有 1 行关系记录，得到 %d", count)
	}
}
