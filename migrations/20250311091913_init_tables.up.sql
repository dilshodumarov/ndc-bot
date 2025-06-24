CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. client_type
CREATE TABLE IF NOT EXISTS "client_type" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "name" VARCHAR NOT NULL UNIQUE,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. role
CREATE TABLE IF NOT EXISTS "role" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "name" VARCHAR(255) NOT NULL UNIQUE,
    "client_type_id" UUID REFERENCES "client_type"("guid"),
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. user
CREATE TABLE IF NOT EXISTS "user" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "first_name" VARCHAR(255) NOT NULL,
    "last_name" VARCHAR(255),
    "email" VARCHAR(255) UNIQUE,
    "phone_number" VARCHAR(255),
    "password" VARCHAR(255),
    "role_id" UUID REFERENCES "role"("guid") ON DELETE CASCADE,
    "is_active" BOOL DEFAULT FALSE,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. business
CREATE TABLE IF NOT EXISTS "business" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "owner_id" UUID REFERENCES "user"("guid") ON DELETE CASCADE, 
    "name" VARCHAR(255),
    "tg_user_name" VARCHAR(255),
    "description" VARCHAR(255),
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP DEFAULT NULL
);

-- 5. integration
CREATE TABLE IF NOT EXISTS "integration" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "owner_id" UUID REFERENCES "business"("guid") ON DELETE CASCADE,
    "integration_token" VARCHAR(255),
    "integration_user_name" VARCHAR(255),
    "status" VARCHAR(10) DEFAULT 'active',
    "started_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "stoped_at" TIMESTAMP,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);
   
CREATE TABLE IF NOT EXISTS "settings" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "name" VARCHAR(50),
    "brand_name" VARCHAR(50),
    "business_name" VARCHAR(50),
    "status" BOOLEAN DEFAULT TRUE,
    "business_id" UUID REFERENCES "business"("guid") ON DELETE CASCADE,
    "prompt_text" TEXT,
    "prompt_order" JSONB,
    "waiting_time" int,
    "prompt_product" TEXT,
    "token_limit" INT DEFAULT 300,
    "chat_token" BIGINT  DEFAULT 300,
    "intelligence_level" INT DEFAULT 100,
    "error_message" TEXT,
    "first_message" TEXT,
    "is_stop" BOOLEAN DEFAULT FALSE,
    "stop_until" int DEFAULT 0,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);

-- 6. category
CREATE TABLE IF NOT EXISTS "category" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "name" VARCHAR(255) NOT NULL UNIQUE,
    "business_id" UUID REFERENCES "business"("guid"),
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. attribute
CREATE TABLE IF NOT EXISTS "attribute" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "name" VARCHAR(255) NOT NULL,
    "category_id" UUID REFERENCES "category"("guid"),
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 8. product
CREATE TABLE IF NOT EXISTS "product" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "product_id" BIGSERIAL UNIQUE,
    "business_id" UUID REFERENCES "business"("guid"),
    "name" VARCHAR(255) NOT NULL,
    "category_id" UUID REFERENCES "category"("guid"),
    "short_info" VARCHAR(255),
    "description" TEXT,
    "status" BOOLEAN DEFAULT TRUE,
    "cost" INT NOT NULL,
    "count" INT NOT NULL,
    "discount_cost" INT,
    "discount" INT,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"  TIMESTAMP
);

-- 9. client
CREATE TABLE IF NOT EXISTS "client" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "client_id" BIGSERIAL UNIQUE,
    "platform_id" VARCHAR(255),
    "first_name" VARCHAR(255) NOT NULL,
    "chat_id" BIGINT,
    "bussnes_id" UUID REFERENCES "business"("guid"),
    "phone" VARCHAR(255),
    "user_name" VARCHAR(50),
    "from_chanel" VARCHAR(50),
    "order_status" VARCHAR(50),
    "location" TEXT,
    "location_text"VARCHAR(50),
    "goal" VARCHAR(50) ,
    "is_block" BOOLEAN DEFAULT FALSE,
    "is_stop"  BOOLEAN DEFAULT FALSE,
    "is_pauze"  BOOLEAN DEFAULT FALSE,
    "stop_until" TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 10. order
CREATE TABLE IF NOT EXISTS "order" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "order_id" BIGSERIAL UNIQUE,
    "order_guid" uuid,
    "client_id" UUID REFERENCES "client"("guid"),
    "business_id" UUID REFERENCES "business"("guid"),
    "location_url" TEXT,
    "location" VARCHAR(55),
    "image_url" TEXT,
    "canceled_reason" TEXT,
    "status" VARCHAR(255),
    "user_note" VARCHAR(55),
    "status_number" int,
    "platform" VARCHAR(50) DEFAULT 'bot',
    "order_status_id" UUID REFERENCES"order_status"("guid"),
    "total_price" NUMERIC(10, 2) NOT NULL,
    "payment_method" VARCHAR(255),
    "status_changed_time" TIMESTAMP,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
);

-- 11. order_products
CREATE TABLE IF NOT EXISTS "order_products" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "order_id" UUID REFERENCES "order"("guid") ON DELETE CASCADE,
    "product_id" UUID REFERENCES "product"("guid") ON DELETE CASCADE,
    "count" INT NOT NULL,
    "price" NUMERIC(10, 2) NOT NULL,
    "total_price"  NUMERIC(10, 2) NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 12. bot_commands
CREATE TABLE IF NOT EXISTS "bot_commands" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "integration_id" UUID NOT NULL REFERENCES "integration"("guid") ON DELETE CASCADE,
    "command" TEXT NOT NULL, 
    "response" TEXT NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 13. chat_history
CREATE TABLE IF NOT EXISTS "chat_history" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "message_id" BIGINT,
    "business_id" UUID REFERENCES "business"("guid"),
    "phone" VARCHAR(15),
    "message" TEXT, 
    "chat_id" BIGINT NOT NULL,
    "platform" VARCHAR(50),
    "platform_id" VARCHAR(255),
    "ai_response" TEXT,
    "reply_to_message_id" BIGINT,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "telegram_accaunt" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "number" VARCHAR(50),
    "user_id" VARCHAR(155),
    "from" VARCHAR(50),  -- 'telegram' yoki 'instagram' deb kiritiladi
    "business_id" UUID REFERENCES "business"("guid"),
    "status" VARCHAR(10) DEFAULT 'active',
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS "notifications" (
    "guid" UUID PRIMARY KEY DEFAULT gen_random_uuid(),   
    "user_id" UUID NOT NULL,                           
    "title" TEXT NOT NULL,                            
    "message" TEXT NOT NULL,                           
    "type" VARCHAR(50) DEFAULT 'info',                 
    "is_read" BOOLEAN DEFAULT FALSE,                  
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT now(), 
    "read_at" TIMESTAMP WITH TIME ZONE                
);

CREATE TABLE IF NOT EXISTS "product_pictures" (
    "guid" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "product_id" UUID NOT NULL REFERENCES "product"("guid") ON DELETE CASCADE,
    "image_url" TEXT NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "client_token_usage" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "business_id" UUID NOT NULL REFERENCES "business"("guid") ON DELETE CASCADE,
    "source_type" VARCHAR(15) NOT NULL,        -- bot / telegram accaunt
    "used_for" VARCHAR(50) NOT NULL,           -- example: 'generate_text', 'image_upload'
    "request_tokens" INT NOT NULL CHECK (request_tokens >= 0),
    "response_tokens" INT NOT NULL CHECK (response_tokens >= 0),
    "total_tokens" INT GENERATED ALWAYS AS (request_tokens + response_tokens) STORED,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "order_status_type" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "name" TEXT NOT NULL UNIQUE, 
    "status_number" BIGSERIAL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "order_status" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "business_id" UUID NOT NULL REFERENCES "business"("guid") ON DELETE CASCADE,
    "type_id" UUID NOT NULL REFERENCES "order_status_type"("guid") ON DELETE CASCADE,
    "custom_name" TEXT NOT NULL, 
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "fon_color"  VARCHAR(50),
    UNIQUE ("business_id", "type_id") 
);


CREATE TABLE IF NOT EXISTS "database" (
    "guid" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "name" VARCHAR(50),
    "description" TEXT,
    "tokens" int,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);




