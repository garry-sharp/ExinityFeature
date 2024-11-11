DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateways') THEN
        CREATE TABLE gateways (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL UNIQUE,
            data_format_supported VARCHAR(50) NOT NULL,  
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'countries') THEN
        CREATE TABLE countries (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL UNIQUE,
            code CHAR(2) NOT NULL UNIQUE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateway_countries') THEN
        CREATE TABLE gateway_countries (
            gateway_id INT NOT NULL, 
            country_id INT NOT NULL,
            PRIMARY KEY (gateway_id, country_id),
            CONSTRAINT fk_gateway FOREIGN KEY (gateway_id) REFERENCES gateways (id) ON DELETE CASCADE,
            CONSTRAINT fk_country FOREIGN KEY (country_id) REFERENCES countries (id) ON DELETE CASCADE
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type
        WHERE typname = 'transaction_status'
    ) THEN
        CREATE TYPE transaction_status AS ENUM ('DRAFT', 'SENT', 'SUCCESS', 'FAILED');
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type
        WHERE typname = 'transaction_type'
    ) THEN
        CREATE TYPE transaction_type AS ENUM ('DEPOSIT', 'WITHDRAWAL');
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'transactions') THEN
        CREATE TABLE transactions (
            id SERIAL PRIMARY KEY,
            amount DECIMAL(10, 2) NOT NULL,
            type transaction_type NOT NULL,
            status transaction_status NOT NULL DEFAULT 'DRAFT',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  
            gateway_id INT NOT NULL,  
            country_id INT NOT NULL,  
            user_id INT NOT NULL
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(255) NOT NULL UNIQUE,
            email VARCHAR(255) NOT NULL UNIQUE,
            country_id INT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'currencies') THEN
        CREATE TABLE currencies (
            id SERIAL PRIMARY KEY,
            symbol CHAR(3) NOT NULL
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'country_currency') THEN
        CREATE TABLE country_currency (
            country_id INT NOT NULL,
            currency_id INT NOT NULL,
            PRIMARY KEY(country_id, currency_id),
            CONSTRAINT fk_country FOREIGN KEY (country_id) REFERENCES countries (id) ON DELETE CASCADE,
            CONSTRAINT fk_currency FOREIGN KEY (currency_id) REFERENCES currencies (id) ON DELETE CASCADE
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.views 
        WHERE table_name = 'gateway_country_currency'
    ) THEN
        EXECUTE 'CREATE VIEW gateway_country_currency AS 
                 SELECT 
                    g.*,
                    co.id as country_id,
                    co.name AS country_name, 
                    cu.id AS currency_id, 
                    cu.symbol AS currency_symbol
                 FROM 
                    gateways g
                    JOIN gateway_countries gc ON gc.gateway_id = g.id
                    JOIN countries co ON gc.country_id = co.id
                    JOIN country_currency cc ON co.id = cc.country_id
                    JOIN currencies cu ON cc.currency_id = cu.id;';
    END IF;
END $$;

