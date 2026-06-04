<script setup lang="ts">
import type {
  ThemeBackgroundEffectType,
  ThemeClickEffectType,
  ThemeCursorEffectType,
  ThemeEffectConfig,
  ThemeEffectScope,
} from '@/types/theme'

const config = defineModel<ThemeEffectConfig>({ required: true })

const scopes: Array<{ value: ThemeEffectScope; label: string }> = [
  { value: 'all', label: '主页和文章' },
  { value: 'home', label: '仅主页' },
  { value: 'post', label: '仅文章' },
]

const backgroundTypes: Array<{ value: ThemeBackgroundEffectType; label: string }> = [
  { value: 'off', label: '关闭' },
  { value: 'floating-shapes', label: '漂浮图形' },
]

const cursorTypes: Array<{ value: ThemeCursorEffectType; label: string }> = [
  { value: 'off', label: '关闭' },
  { value: 'particle-trail', label: '粒子轨迹' },
]

const clickTypes: Array<{ value: ThemeClickEffectType; label: string }> = [
  { value: 'off', label: '关闭' },
  { value: 'heart', label: '爱心' },
  { value: 'star', label: '五角星' },
  { value: 'dot', label: '圆点' },
]
</script>

<template>
  <section class="space-y-5">
    <div>
      <h2 class="text-base font-semibold text-slate-700">页面特效</h2>
      <p class="text-xs text-slate-400 mt-1">使用受限 Canvas 绘制；系统开启“减少动态效果”时会自动停用。</p>
    </div>

    <label class="effect-row">
      <span>
        <strong>生效范围</strong>
        <small>选择特效出现的页面</small>
      </span>
      <select v-model="config.scope" class="effect-select">
        <option v-for="scope in scopes" :key="scope.value" :value="scope.value">{{ scope.label }}</option>
      </select>
    </label>

    <div class="effect-card">
      <label class="effect-row">
        <span>
          <strong>背景特效</strong>
          <small>缓慢漂浮的圆形、方形与三角形</small>
        </span>
        <select v-model="config.background.type" class="effect-select">
          <option v-for="item in backgroundTypes" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
      </label>
      <div v-if="config.background.type !== 'off'" class="effect-controls">
        <label>颜色 <input v-model="config.background.color" type="color" /></label>
        <label>数量 {{ config.background.count }} <input v-model.number="config.background.count" type="range" min="1" max="60" /></label>
        <label>大小 {{ config.background.size }} <input v-model.number="config.background.size" type="range" min="4" max="80" /></label>
        <label>透明度 {{ config.background.opacity.toFixed(2) }} <input v-model.number="config.background.opacity" type="range" min="0.05" max="1" step="0.05" /></label>
        <label>速度 {{ config.background.speed.toFixed(1) }} <input v-model.number="config.background.speed" type="range" min="0.1" max="2" step="0.1" /></label>
      </div>
    </div>

    <div class="effect-card">
      <label class="effect-row">
        <span>
          <strong>鼠标跟随</strong>
          <small>指针移动时留下轻量粒子轨迹</small>
        </span>
        <select v-model="config.cursor.type" class="effect-select">
          <option v-for="item in cursorTypes" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
      </label>
      <div v-if="config.cursor.type !== 'off'" class="effect-controls">
        <label>颜色 <input v-model="config.cursor.color" type="color" /></label>
        <label>粒子上限 {{ config.cursor.count }} <input v-model.number="config.cursor.count" type="range" min="1" max="50" /></label>
        <label>大小 {{ config.cursor.size }} <input v-model.number="config.cursor.size" type="range" min="1" max="20" /></label>
        <label>透明度 {{ config.cursor.opacity.toFixed(2) }} <input v-model.number="config.cursor.opacity" type="range" min="0.05" max="1" step="0.05" /></label>
        <label>时长 {{ config.cursor.duration }}ms <input v-model.number="config.cursor.duration" type="range" min="100" max="2000" step="50" /></label>
      </div>
    </div>

    <div class="effect-card">
      <label class="effect-row">
        <span>
          <strong>点击特效</strong>
          <small>点击页面时弹出可信预设图形</small>
        </span>
        <select v-model="config.click.type" class="effect-select">
          <option v-for="item in clickTypes" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
      </label>
      <div v-if="config.click.type !== 'off'" class="effect-controls">
        <label>颜色 <input v-model="config.click.color" type="color" /></label>
        <label>数量 {{ config.click.count }} <input v-model.number="config.click.count" type="range" min="1" max="24" /></label>
        <label>大小 {{ config.click.size }} <input v-model.number="config.click.size" type="range" min="4" max="60" /></label>
        <label>透明度 {{ config.click.opacity.toFixed(2) }} <input v-model.number="config.click.opacity" type="range" min="0.05" max="1" step="0.05" /></label>
        <label>时长 {{ config.click.duration }}ms <input v-model.number="config.click.duration" type="range" min="200" max="3000" step="50" /></label>
      </div>
    </div>
  </section>
</template>

<style scoped>
.effect-card {
  padding: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  background: rgb(248 250 252 / 0.6);
}

.effect-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.effect-row span {
  display: flex;
  flex-direction: column;
}

.effect-row strong {
  color: #334155;
  font-size: 0.875rem;
  font-weight: 500;
}

.effect-row small {
  margin-top: 0.125rem;
  color: #94a3b8;
  font-size: 0.75rem;
}

.effect-select {
  min-width: 8rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  outline: none;
  background: white;
  color: #475569;
  font-size: 0.875rem;
}

.effect-controls {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem 1.25rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #e2e8f0;
}

.effect-controls label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #64748b;
  font-size: 0.75rem;
}

.effect-controls input[type='range'] {
  min-width: 0;
  flex: 1;
  accent-color: #6366f1;
}

.effect-controls input[type='color'] {
  width: 2.5rem;
  height: 1.75rem;
  padding: 0.125rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  background: white;
  cursor: pointer;
}

@media (max-width: 720px) {
  .effect-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .effect-controls {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
