-- +goose Up
CREATE TABLE email_verify_codes (
  id         BIGSERIAL    PRIMARY KEY,
  email      VARCHAR(120) NOT NULL,
  code       VARCHAR(10)  NOT NULL,
  expires_at TIMESTAMPTZ  NOT NULL,
  used_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_evc_email ON email_verify_codes(email);

CREATE TABLE invite_codes (
  id         BIGSERIAL   PRIMARY KEY,
  code       VARCHAR(20) NOT NULL UNIQUE,
  created_by BIGINT      NOT NULL REFERENCES users(id),
  is_used    BOOLEAN     NOT NULL DEFAULT false,
  used_by    BIGINT      REFERENCES users(id),
  used_at    TIMESTAMPTZ,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS invite_codes;
DROP TABLE IF EXISTS email_verify_codes;
