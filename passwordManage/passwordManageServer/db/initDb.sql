CREATE TABLE IF NOT EXISTS input_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- 主键
		app_name TEXT,		-- 应用名称
		is_app_name_encrypted BOOLEAN, -- 是否加密
		username TEXT,		-- 用户名
		is_username_encrypted BOOLEAN,		-- 是否加密
		input_type TEXT,	-- 输入类型
		password TEXT,		-- 密码
		key_file TEXT,		-- 密钥文件
		url TEXT,			-- 网址
		is_url_encrypted BOOLEAN,	-- 是否加密
		notes TEXT,			-- 备注
		is_notes_encrypted BOOLEAN,	-- 是否加密
		tags TEXT,			-- 标签
		is_tags_encrypted BOOLEAN,	-- 是否加密
		strength TEXT,		-- 强度
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,	-- 创建时间
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,	-- 更新时间
		chose_encrypt TEXT,
		key TEXT,
		count INTEGER
		
	);