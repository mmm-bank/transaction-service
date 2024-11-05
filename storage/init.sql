CREATE EXTENSION amqp;

CREATE TABLE accounts (
    account_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    account_number BYTEA NOT NULL,
    phone_number BYTEA NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO amqp.broker (host, port, vhost, username, password)
VALUES ('rabbitmq', 5672, '/', 'guest', 'guest');

CREATE OR REPLACE PROCEDURE process_transfer(
    transaction_id UUID,
    expected_sender_id UUID,
    sender_account_id UUID,
    receiver_account_id UUID,
    amount BIGINT
) LANGUAGE plpgsql AS $$
DECLARE
    sender_balance BIGINT;
    sender_id UUID;
    receiver_balance BIGINT;
    receiver_id UUID;
BEGIN
    SELECT balance, user_id
    INTO sender_balance, sender_id
    FROM accounts
    WHERE account_id = sender_account_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'sender account not found %', sender_account_id;
    END IF;

    IF sender_id != expected_sender_id THEN
       RAISE EXCEPTION 'access denied; wrong user %', sender_id;
    END IF;

    IF sender_balance < amount THEN
        RAISE EXCEPTION 'insufficient funds on account %', sender_account_id;
    END IF;

    SELECT balance, user_id
    INTO receiver_balance, receiver_id
    FROM accounts
    WHERE account_id = receiver_account_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'receiver account not found %', receiver_account_id;
    END IF;

    UPDATE accounts
    SET balance = balance - amount
    WHERE account_id = sender_account_id;

    UPDATE accounts
    SET balance = balance + amount
    WHERE account_id = receiver_account_id;

    PERFORM amqp.publish(
        1,
        'transaction_events',
        '',
        json_build_object(
            'transaction_id',  transaction_id,
            'sender_id', sender_id,
            'sender_account_id', sender_account_id,
            'receiver_id',  receiver_id,
            'receiver_account_id', receiver_account_id,
            'amount', amount
        )::text
    );
END;
$$;

CREATE TABLE card_to_account (
    card_number BYTEA PRIMARY KEY,
    account_id  UUID NOT NULL,
    linked_at   TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (account_id) REFERENCES accounts (account_id) ON DELETE CASCADE
);

CREATE TABLE phone_to_account (
    phone_number BYTEA PRIMARY KEY,
    account_id  UUID NOT NULL,
    linked_at   TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (account_id) REFERENCES accounts (account_id) ON DELETE CASCADE
);