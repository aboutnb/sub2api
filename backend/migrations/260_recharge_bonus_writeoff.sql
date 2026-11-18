-- No historical backfill. Monetary amounts use the billing ledger's scale.
CREATE TABLE recharge_bonus_writeoff_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id),
    enabled BOOLEAN NOT NULL DEFAULT false,
    effective_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    enabled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    amount_limit NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (amount_limit >= 0),
    written_off NUMERIC(20,8) NOT NULL DEFAULT 0 CHECK (written_off >= 0),
    consent_note TEXT NOT NULL DEFAULT '',
    updated_by BIGINT NOT NULL,
    revision BIGINT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE recharge_bonus_writeoff_batches (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    order_id BIGINT NOT NULL UNIQUE REFERENCES payment_orders(id),
    principal NUMERIC(20,8) NOT NULL CHECK (principal > 0),
    bonus NUMERIC(20,8) NOT NULL CHECK (bonus > 0),
    principal_used NUMERIC(20,8) NOT NULL DEFAULT 0,
    bonus_used NUMERIC(20,8) NOT NULL DEFAULT 0,
    written_off NUMERIC(20,8) NOT NULL DEFAULT 0,
    policy_revision BIGINT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (principal_used BETWEEN 0 AND principal),
    CHECK (bonus_used >= 0 AND written_off >= 0 AND bonus_used + written_off <= bonus)
);
CREATE INDEX recharge_bonus_writeoff_batches_user_idx
    ON recharge_bonus_writeoff_batches(user_id, id) WHERE active;
-- Capture the policy at order insertion, so clock skew cannot enroll an old order.
CREATE FUNCTION snapshot_recharge_bonus_policy() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE revision BIGINT;
BEGIN
    NEW.provider_snapshot := COALESCE(NEW.provider_snapshot,'{}'::jsonb)-'bonus_writeoff_revision';
    SELECT s.revision INTO revision FROM recharge_bonus_writeoff_settings s
      WHERE s.user_id=NEW.user_id AND s.enabled AND NOW()>=s.effective_at;
    IF FOUND THEN
      NEW.provider_snapshot := COALESCE(NEW.provider_snapshot,'{}'::jsonb)
        || jsonb_build_object('bonus_writeoff_revision',revision);
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER recharge_bonus_policy_snapshot BEFORE INSERT ON payment_orders
    FOR EACH ROW EXECUTE FUNCTION snapshot_recharge_bonus_policy();
CREATE FUNCTION register_recharge_bonus_batch(uid BIGINT, code TEXT) RETURNS VOID LANGUAGE sql AS $$
 INSERT INTO recharge_bonus_writeoff_batches(user_id,order_id,principal,bonus,policy_revision)
 SELECT o.user_id,o.id,o.amount-o.bonus_amount,o.bonus_amount,s.revision
 FROM payment_orders o JOIN recharge_bonus_writeoff_settings s ON s.user_id=o.user_id
 WHERE o.user_id=uid AND o.recharge_code=code AND o.order_type='balance'
 AND o.provider_snapshot->>'recharge_bonus_mode'='bonus'
 AND o.provider_snapshot->>'bonus_writeoff_revision'=s.revision::text
 AND o.amount>o.bonus_amount AND o.bonus_amount>0 AND s.enabled
 AND NOW()>=GREATEST(s.enabled_at,s.effective_at)
 ON CONFLICT(order_id) DO NOTHING;
$$;
CREATE TABLE recharge_bonus_writeoff_audit (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    batch_id BIGINT REFERENCES recharge_bonus_writeoff_batches(id),
    request_id TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL,
    amount NUMERIC(20,8) NOT NULL DEFAULT 0,
    balance_after NUMERIC(20,8),
    actor_id BIGINT,
    detail JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX recharge_bonus_writeoff_audit_user_idx ON recharge_bonus_writeoff_audit(user_id, id DESC);

CREATE TABLE recharge_bonus_writeoff_holds (
    user_id BIGINT NOT NULL REFERENCES users(id),
    hold_key TEXT NOT NULL,
    batch_id BIGINT REFERENCES recharge_bonus_writeoff_batches(id),
    principal NUMERIC(20,8) NOT NULL DEFAULT 0,
    bonus NUMERIC(20,8) NOT NULL DEFAULT 0,
    liability NUMERIC(20,8) NOT NULL DEFAULT 0
);
CREATE INDEX recharge_bonus_writeoff_holds_user_idx ON recharge_bonus_writeoff_holds(user_id,hold_key);

-- Called after moving available balance to frozen balance, under the user lock.
CREATE FUNCTION reserve_recharge_bonus_writeoff(uid BIGINT, cost NUMERIC, hold_key TEXT)
RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE b RECORD; available NUMERIC; tracked NUMERIC; remaining NUMERIC := cost; p NUMERIC; g NUMERIC;
    liability NUMERIC; total NUMERIC := 0; cfg recharge_bonus_writeoff_settings%ROWTYPE;
BEGIN
    SELECT balance+cost INTO available FROM users WHERE id=uid FOR UPDATE;
    SELECT * INTO cfg FROM recharge_bonus_writeoff_settings WHERE user_id=uid;
    SELECT COALESCE(SUM(principal-principal_used+bonus-bonus_used-written_off),0)
      - COALESCE((SELECT SUM(h.principal+h.bonus+h.liability) FROM recharge_bonus_writeoff_holds h
          JOIN recharge_bonus_writeoff_batches x ON x.id=h.batch_id WHERE h.user_id=uid AND x.active),0)
      INTO tracked FROM recharge_bonus_writeoff_batches WHERE user_id=uid AND active;
    p := LEAST(remaining,GREATEST(0,available-tracked));
    INSERT INTO recharge_bonus_writeoff_holds VALUES(uid,hold_key,NULL,p,0,0);
    remaining := remaining-p;
    FOR b IN SELECT x.*,
      COALESCE((SELECT SUM(h.principal) FROM recharge_bonus_writeoff_holds h WHERE h.batch_id=x.id),0) AS held_p,
      COALESCE((SELECT SUM(h.bonus+h.liability) FROM recharge_bonus_writeoff_holds h WHERE h.batch_id=x.id),0) AS held_b
      FROM recharge_bonus_writeoff_batches x WHERE user_id=uid AND active ORDER BY id LOOP
      EXIT WHEN remaining<=0;
      p := LEAST(remaining,GREATEST(0,b.principal-b.principal_used-b.held_p));
      remaining := remaining-p;
      liability := 0;
      IF cfg.enabled AND NOW()>=cfg.effective_at THEN
        liability := LEAST(CEIL(p*b.bonus/b.principal*100000000)/100000000,
          GREATEST(0,b.bonus-b.bonus_used-b.written_off-b.held_b));
        IF cfg.amount_limit>0 THEN liability := LEAST(liability,GREATEST(0,cfg.amount_limit-cfg.written_off-total)); END IF;
      END IF;
      total := total+liability;
      g := LEAST(remaining,GREATEST(0,b.bonus-b.bonus_used-b.written_off-b.held_b-liability));
      remaining := remaining-g;
      INSERT INTO recharge_bonus_writeoff_holds VALUES(uid,hold_key,b.id,p,g,liability);
    END LOOP;
    IF remaining>0 OR available-cost<total THEN RAISE EXCEPTION 'bonus hold insufficient balance'; END IF;
    UPDATE users SET balance=balance-total,frozen_balance=COALESCE(frozen_balance,0)+total WHERE id=uid;
END $$;

CREATE FUNCTION release_recharge_bonus_liability(uid BIGINT, key TEXT) RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE amount NUMERIC;
BEGIN
    PERFORM id FROM users WHERE id=uid FOR UPDATE;
    SELECT COALESCE(SUM(liability),0) INTO amount FROM recharge_bonus_writeoff_holds WHERE user_id=uid AND hold_key=key;
    UPDATE users SET balance=balance+amount,frozen_balance=COALESCE(frozen_balance,0)-amount WHERE id=uid;
    UPDATE recharge_bonus_writeoff_holds SET liability=0 WHERE user_id=uid AND hold_key=key;
END $$;

-- Caller already holds the user row lock, and has applied the normal debit.
-- Held funds are excluded; capture consumes only the original hold allocation.
CREATE FUNCTION apply_recharge_bonus_writeoff(
    uid BIGINT, cost NUMERIC, request_key TEXT, capture_key TEXT DEFAULT ''
) RETURNS NUMERIC LANGUAGE plpgsql AS $$
DECLARE
    cfg recharge_bonus_writeoff_settings%ROWTYPE;
    b recharge_bonus_writeoff_batches%ROWTYPE;
    available NUMERIC;
    frozen NUMERIC;
    tracked NUMERIC;
    remaining NUMERIC;
    take NUMERIC;
    extra NUMERIC;
    total NUMERIC := 0;
    held_p NUMERIC;
    held_b NUMERIC;
    own_p NUMERIC;
    own_b NUMERIC;
BEGIN
    IF cost <= 0 THEN
      DELETE FROM recharge_bonus_writeoff_holds h WHERE h.user_id=uid AND h.hold_key=capture_key;
      RETURN 0;
    END IF;
    SELECT balance, COALESCE(frozen_balance,0) INTO available, frozen
      FROM users WHERE id=uid FOR UPDATE;
    SELECT * INTO cfg FROM recharge_bonus_writeoff_settings WHERE user_id=uid;
    IF NOT FOUND THEN
      DELETE FROM recharge_bonus_writeoff_holds h WHERE h.user_id=uid AND h.hold_key=capture_key;
      RETURN 0;
    END IF;
    SELECT COALESCE(SUM(principal-principal_used+bonus-bonus_used-written_off),0)
      INTO tracked FROM recharge_bonus_writeoff_batches WHERE user_id=uid AND active;
    SELECT tracked-COALESCE(SUM(h.principal+h.bonus+h.liability),0) INTO tracked
      FROM recharge_bonus_writeoff_holds h JOIN recharge_bonus_writeoff_batches x ON x.id=h.batch_id
      WHERE h.user_id=uid AND x.active;
    IF capture_key<>'' THEN
      IF NOT EXISTS(SELECT 1 FROM recharge_bonus_writeoff_holds h WHERE h.user_id=uid AND h.hold_key=capture_key) THEN RETURN 0; END IF;
      SELECT GREATEST(0,cost-COALESCE(SUM(h.principal+h.bonus),0)) INTO remaining
        FROM recharge_bonus_writeoff_holds h LEFT JOIN recharge_bonus_writeoff_batches x ON x.id=h.batch_id
        WHERE h.user_id=uid AND h.hold_key=capture_key AND (h.batch_id IS NULL OR NOT x.active);
    ELSE
      remaining := GREATEST(0, cost - GREATEST(0, available+cost-tracked));
    END IF;
    FOR b IN SELECT * FROM recharge_bonus_writeoff_batches
      WHERE user_id=uid AND active ORDER BY id FOR UPDATE LOOP
        EXIT WHEN remaining <= 0;
        SELECT COALESCE(SUM(h.principal),0),COALESCE(SUM(h.bonus+h.liability),0),
          COALESCE(SUM(h.principal) FILTER(WHERE h.hold_key=capture_key),0),
          COALESCE(SUM(h.bonus) FILTER(WHERE h.hold_key=capture_key),0)
          INTO held_p,held_b,own_p,own_b FROM recharge_bonus_writeoff_holds h WHERE h.batch_id=b.id;
        take := LEAST(remaining,CASE WHEN capture_key<>'' THEN own_p ELSE GREATEST(0,b.principal-b.principal_used-held_p) END);
        b.principal_used := b.principal_used+take;
        remaining := remaining-take;
        extra := 0;
        IF cfg.enabled AND NOW() >= cfg.effective_at AND take > 0 THEN
            -- Only this consumption's cumulative increment: never catch up skipped debits.
            extra := ROUND(b.principal_used*b.bonus/b.principal,8)
                   - ROUND((b.principal_used-take)*b.bonus/b.principal,8);
            extra := LEAST(extra, GREATEST(0,b.bonus-b.bonus_used-b.written_off-held_b), GREATEST(available-total,0));
            IF cfg.amount_limit > 0 THEN
                extra := LEAST(extra, GREATEST(cfg.amount_limit-cfg.written_off-total,0));
            END IF;
        END IF;
        b.written_off := b.written_off+extra;
        total := total+extra;
        take := LEAST(remaining,CASE WHEN capture_key<>'' THEN own_b ELSE GREATEST(0,b.bonus-b.bonus_used-b.written_off-held_b) END);
        b.bonus_used := b.bonus_used+take;
        remaining := remaining-take;
        UPDATE recharge_bonus_writeoff_batches SET principal_used=b.principal_used,
          bonus_used=b.bonus_used, written_off=b.written_off WHERE id=b.id;
        IF extra > 0 THEN
            INSERT INTO recharge_bonus_writeoff_audit(user_id,batch_id,request_id,action,amount,balance_after,detail)
            VALUES(uid,b.id,request_key,'writeoff',extra,available-total,
              jsonb_build_object('principal',b.principal,'bonus',b.bonus,'policy_revision',cfg.revision,
                'order_id',b.order_id,'balance_before',available-total+extra));
        END IF;
    END LOOP;
    IF capture_key<>'' THEN DELETE FROM recharge_bonus_writeoff_holds h WHERE h.user_id=uid AND h.hold_key=capture_key; END IF;
    IF total > 0 THEN
        UPDATE users SET balance=balance-total WHERE id=uid;
        UPDATE recharge_bonus_writeoff_settings SET written_off=written_off+total WHERE user_id=uid;
    END IF;
    RETURN total;
END $$;

-- Unclassified balance reductions (admin adjustments, other purchases) must not
-- leave stale earmarked funds that could later debit unrelated principal.
CREATE FUNCTION retire_bonus_batches_on_adjustment() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.balance+COALESCE(NEW.frozen_balance,0) < OLD.balance+COALESCE(OLD.frozen_balance,0)
       AND COALESCE(current_setting('aivoza.bonus_usage',true),'') <> 'on' THEN
        WITH retired AS (
          UPDATE recharge_bonus_writeoff_batches SET active=false WHERE user_id=NEW.id AND active RETURNING id
        ) INSERT INTO recharge_bonus_writeoff_audit(user_id,batch_id,action,detail)
          SELECT NEW.id,id,'adjustment_retired',jsonb_build_object('reason','non_usage_balance_reduction') FROM retired;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER bonus_writeoff_balance_adjustment AFTER UPDATE OF balance,frozen_balance ON users
    FOR EACH ROW EXECUTE FUNCTION retire_bonus_batches_on_adjustment();

-- Refund attempts permanently retire the affected batch and return previously
-- written-off bonus once, before the existing refund workflow executes its debit.
CREATE FUNCTION retire_bonus_batch_for_refund() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE b recharge_bonus_writeoff_batches%ROWTYPE;
BEGIN
    IF NEW.status='REFUNDING' AND OLD.status IS DISTINCT FROM 'REFUNDING' THEN
      PERFORM id FROM users WHERE id=NEW.user_id FOR UPDATE;
      SELECT * INTO b FROM recharge_bonus_writeoff_batches WHERE order_id=NEW.id FOR UPDATE;
      IF FOUND AND NOT EXISTS(SELECT 1 FROM recharge_bonus_writeoff_audit WHERE batch_id=b.id AND action='refund_restore') THEN
        UPDATE recharge_bonus_writeoff_batches SET active=false WHERE id=b.id;
        UPDATE users SET balance=balance+b.written_off WHERE id=NEW.user_id;
        INSERT INTO recharge_bonus_writeoff_audit(user_id,batch_id,action,amount,detail)
        VALUES(NEW.user_id,b.id,'refund_restore',-b.written_off,jsonb_build_object('order_id',NEW.id));
      END IF;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER bonus_writeoff_refund BEFORE UPDATE OF status ON payment_orders
    FOR EACH ROW EXECUTE FUNCTION retire_bonus_batch_for_refund();
