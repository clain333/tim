CREATE TABLE IF NOT EXISTS user (
id BIGINT UNSIGNED  PRIMARY KEY,
name VARCHAR(64) NOT NULL,
password VARCHAR(128) NOT NULL,
email VARCHAR(64) UNIQUE,
avatar VARCHAR(255),
introduction TEXT,
create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
INDEX idx_email (email),
INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

CREATE TABLE IF NOT EXISTS device (
id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
user_id BIGINT UNSIGNED NOT NULL,
ip VARCHAR(45) NOT NULL,
os VARCHAR(50),
device_tag VARCHAR(255) NOT NULL,
login_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
INDEX idx_user_id (user_id),
CONSTRAINT fk_device_user
FOREIGN KEY (user_id) REFERENCES user(id)
ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设备登录记录表';

CREATE TABLE IF NOT EXISTS friend (
id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
conversation_id BIGINT UNSIGNED NOT NULL,
min_id BIGINT UNSIGNED NOT NULL,
max_id BIGINT UNSIGNED NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

UNIQUE KEY uk_min_max (min_id, max_id),
INDEX idx_min_id (min_id),
INDEX idx_max_id (max_id),

CONSTRAINT fk_friend_user_min
FOREIGN KEY (min_id) REFERENCES user(id)
ON DELETE CASCADE,

CONSTRAINT fk_friend_user_max
FOREIGN KEY (max_id) REFERENCES user(id)
ON DELETE CASCADE

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友关系表';

CREATE TABLE IF NOT EXISTS chat_group (
id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
number VARCHAR(50) NOT NULL UNIQUE,  -- 群号唯一
conversation_id BIGINT UNSIGNED NOT NULL,
name VARCHAR(50) NOT NULL,
group_owner BIGINT UNSIGNED NOT NULL,  -- 群主
avatar_url VARCHAR(255),
description TEXT,
member_count BIGINT UNSIGNED DEFAULT 0,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
INDEX idx_number (number),
INDEX idx_owner (group_owner),

CONSTRAINT fk_group_owner
FOREIGN KEY (group_owner) REFERENCES user(id)
ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群聊表';


-- 4. 群成员表
CREATE TABLE IF NOT EXISTS group_member (
id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,  -- 加主键
group_id BIGINT UNSIGNED NOT NULL,
user_id BIGINT UNSIGNED NOT NULL,
role TINYINT UNSIGNED NOT NULL DEFAULT 0,  -- 0=普通成员, 1=群主
joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
UNIQUE KEY uk_group_user (group_id, user_id),
INDEX idx_user_id (user_id),
INDEX idx_group_id (group_id),

CONSTRAINT fk_group_member_group
FOREIGN KEY (group_id) REFERENCES chat_group(id)
ON DELETE CASCADE,

CONSTRAINT fk_group_member_user
FOREIGN KEY (user_id) REFERENCES user(id)
ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群成员表';

-- 6. 好友请求表
CREATE TABLE IF NOT EXISTS friend_request (
id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
send_id BIGINT UNSIGNED NOT NULL,      -- 发起者
receive_id BIGINT UNSIGNED NOT NULL,   -- 接收者（原 revise_id → receive_id）
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

UNIQUE KEY uk_send_receive (send_id, receive_id),  -- 防止重复请求
INDEX idx_send_id (send_id),
INDEX idx_receive_id (receive_id),

CONSTRAINT fk_friend_request_sender
FOREIGN KEY (send_id) REFERENCES user(id)
ON DELETE CASCADE,

CONSTRAINT fk_friend_request_receiver
FOREIGN KEY (receive_id) REFERENCES user(id)
ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友请求表';


CREATE TABLE IF NOT EXISTS group_join_request (
id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
group_id BIGINT UNSIGNED NOT NULL,
user_id BIGINT UNSIGNED NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
UNIQUE KEY uk_group_user (group_id, user_id),
INDEX idx_group_id (group_id),
INDEX idx_user_id (user_id),
CONSTRAINT fk_join_request_group
FOREIGN KEY (group_id) REFERENCES chat_group(id)
ON DELETE CASCADE,
CONSTRAINT fk_join_request_user
FOREIGN KEY (user_id) REFERENCES user(id)
ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群加入请求表';


CREATE TABLE IF NOT EXISTS files (
id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
real_file_name VARCHAR(64) NOT NULL COMMENT '真实文件ID，通常用UUID',
original_filename VARCHAR(255) NOT NULL COMMENT '用户上传时的文件名',
file_hash VARCHAR(64) NOT NULL COMMENT '文件哈希值（如SHA-256）',
file_size BIGINT UNSIGNED NOT NULL COMMENT '文件大小，单位字节',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件信息表';

CREATE TABLE IF NOT EXISTS offline (
device_id BIGINT UNSIGNED NOT NULL,
seq BIGINT UNSIGNED,
# 系统信息为{用户id}
conversation_id BIGINT UNSIGNED,
CONSTRAINT offline_device_id
FOREIGN KEY (device_id) REFERENCES device(id)
ON DELETE CASCADE,
UNIQUE (device_id,conversation_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='离线信息记录表';

CREATE TABLE IF NOT EXISTS new_seq (
# 系统信息为{用户id}
conversation_id BIGINT UNSIGNED NOT NULL,
seq BIGINT UNSIGNED DEFAULT 0,
    INDEX (conversation_id),
    UNIQUE (conversation_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话seq记录';
