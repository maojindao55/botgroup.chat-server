package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// LLMProvider 定义LLM提供商的配置结构
type LLMProvider struct {
	APIKey  string
	BaseURL string
}

// LLMGroup 定义LLM组的配置结构
type LLMGroup struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	Members               []string `json:"members"`
	IsGroupDiscussionMode bool     `json:"isGroupDiscussionMode"`
}

// LLMCharacter 定义LLM角色的配置结构
type LLMCharacter struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Personality  string   `json:"personality"`
	Model        string   `json:"model"`
	Avatar       string   `json:"avatar"`
	CustomPrompt string   `mapstructure:"custom_prompt" json:"custom_prompt"`
	Tags         []string `json:"tags"`
	RAG          bool     `json:"rag"`
	Knowledge    string   `json:"knowledge"`
}

// AliyunSMSConfig 阿里云短信配置结构
type AliyunSMSConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id" json:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" json:"access_key_secret"`
	SignName        string `mapstructure:"sign_name" json:"sign_name"`
	TemplateCode    string `mapstructure:"template_code" json:"template_code"`
}

// RedisConfig Redis配置结构
type RedisConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     string `mapstructure:"port" json:"port"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}

// CloudflareConfig Cloudflare配置结构
type CloudflareConfig struct {
	AccountID   string `mapstructure:"account_id" json:"account_id"`
	APIToken    string `mapstructure:"api_token" json:"api_token"`
	ImagePrefix string `mapstructure:"image_prefix" json:"image_prefix"`
}

// WechatConfig 微信公众号配置结构
type WechatConfig struct {
	AppID            string `mapstructure:"app_id" json:"app_id"`
	AppSecret        string `mapstructure:"app_secret" json:"app_secret"`
	Token            string `mapstructure:"token" json:"token"`
	CallbackURL      string `mapstructure:"callback_url" json:"callback_url"`
	QRExpiresIn      int    `mapstructure:"qr_expires_in" json:"qr_expires_in"`
	SessionExpiresIn int    `mapstructure:"session_expires_in" json:"session_expires_in"`
}

// WebSocketConfig WebSocket配置结构
type WebSocketConfig struct {
	ReadBufferSize  int  `mapstructure:"read_buffer_size" json:"read_buffer_size"`
	WriteBufferSize int  `mapstructure:"write_buffer_size" json:"write_buffer_size"`
	CheckOrigin     bool `mapstructure:"check_origin" json:"check_origin"`
}

// Config 应用配置结构
type Config struct {
	Server struct {
		Port string
	}
	Database struct {
		DSN string
	}
	LLMSystemPrompt string                 `mapstructure:"llm_system_prompt" json:"llm_system_prompt"`
	LLMProviders    map[string]LLMProvider `mapstructure:"llm_providers"`
	LLMModels       map[string]string      `mapstructure:"llm_models"`
	LLMGroups       []*LLMGroup            `mapstructure:"llm_groups"`
	LLMCharacters   []*LLMCharacter        `mapstructure:"llm_characters"`
	SMS             AliyunSMSConfig        `mapstructure:"sms" json:"sms"`
	Redis           RedisConfig            `mapstructure:"redis" json:"redis"`
	JWTSecret       string                 `mapstructure:"jwt_secret" json:"jwt_secret"`
	AuthAccess      int                    `mapstructure:"auth_access" json:"auth_access"`
	ChatRateLimit   int                    `mapstructure:"chat_rate_limit" json:"chat_rate_limit"`
	Cloudflare      CloudflareConfig       `mapstructure:"cloudflare" json:"cloudflare"`
	Wechat          WechatConfig           `mapstructure:"wechat" json:"wechat"`
	WebSocket       WebSocketConfig        `mapstructure:"websocket" json:"websocket"`
}

var AppConfig Config

// LoadConfig 加载配置文件
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./src/config")

	// 设置默认值
	viper.SetDefault("server.port", "8080")

	// 设置环境变量自动绑定
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定具体的环境变量
	viper.BindEnv("sms.access_key_id", "ALIYUN_SMS_ACCESS_KEY_ID")
	viper.BindEnv("sms.access_key_secret", "ALIYUN_SMS_ACCESS_KEY_SECRET")
	viper.BindEnv("sms.sign_name", "ALIYUN_SMS_SIGN_NAME")
	viper.BindEnv("sms.template_code", "ALIYUN_SMS_TEMPLATE_CODE")

	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	viper.BindEnv("jwt_secret", "JWT_SECRET")
	viper.BindEnv("cloudflare.account_id", "CF_ACCOUNT_ID")
	viper.BindEnv("cloudflare.api_token", "CF_API_TOKEN")
	viper.BindEnv("cloudflare.image_prefix", "CF_IMAGE_PREFIX")

	// 微信公众号配置环境变量绑定
	viper.BindEnv("wechat.app_id", "WECHAT_APP_ID")
	viper.BindEnv("wechat.app_secret", "WECHAT_APP_SECRET")
	viper.BindEnv("wechat.token", "WECHAT_TOKEN")
	viper.BindEnv("wechat.callback_url", "WECHAT_CALLBACK_URL")
	viper.BindEnv("wechat.qr_expires_in", "WECHAT_QR_EXPIRES_IN")
	viper.BindEnv("wechat.session_expires_in", "WECHAT_SESSION_EXPIRES_IN")

	// WebSocket配置环境变量绑定
	viper.BindEnv("websocket.read_buffer_size", "WS_READ_BUFFER_SIZE")
	viper.BindEnv("websocket.write_buffer_size", "WS_WRITE_BUFFER_SIZE")
	viper.BindEnv("websocket.check_origin", "WS_CHECK_ORIGIN")

	//是否登录检测
	viper.BindEnv("auth_access", "AUTH_ACCESS")

	//Chat接口限流配置
	viper.BindEnv("chat_rate_limit", "CHAT_RATE_LIMIT")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("未找到配置文件，使用默认配置")
		} else {
			log.Fatalf("读取配置文件错误: %v", err)
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("解析配置文件错误: %v", err)
	}

	// 环境变量覆盖（保留原有逻辑作为后备）
	if port := os.Getenv("SERVER_PORT"); port != "" {
		AppConfig.Server.Port = port
	}

	// 调试日志：检查环境变量读取情况
	log.Printf("SMS配置: AccessKeyID=%s, SignName=%s, TemplateCode=%s",
		AppConfig.SMS.AccessKeyID, AppConfig.SMS.SignName, AppConfig.SMS.TemplateCode)
	log.Printf("Redis配置: Host=%s, Port=%s, Password=%s, DB=%d",
		AppConfig.Redis.Host, AppConfig.Redis.Port, AppConfig.Redis.Password, AppConfig.Redis.DB)
	log.Printf("JWT Secret: %s", AppConfig.JWTSecret)
	log.Printf("MySQL配置: %s", AppConfig.Database.DSN)
	log.Printf("Cloudflare配置: AccountID=%s, APIToken=%s", AppConfig.Cloudflare.AccountID, AppConfig.Cloudflare.APIToken)
	log.Printf("微信配置: AppID=%s, Token=%s, CallbackURL=%s, QRExpiresIn=%d, SessionExpiresIn=%d",
		AppConfig.Wechat.AppID, AppConfig.Wechat.Token, AppConfig.Wechat.CallbackURL,
		AppConfig.Wechat.QRExpiresIn, AppConfig.Wechat.SessionExpiresIn)
	log.Printf("WebSocket配置: ReadBufferSize=%d, WriteBufferSize=%d, CheckOrigin=%t",
		AppConfig.WebSocket.ReadBufferSize, AppConfig.WebSocket.WriteBufferSize, AppConfig.WebSocket.CheckOrigin)

	log.Println("AppConfig:", AppConfig.LLMModels)

	// 确保配置键名称正确映射
	if len(AppConfig.LLMModels) == 0 {
		log.Println("LLMModels为空，尝试手动加载")
		modelsMap := viper.GetStringMapString("llm_models")
		if len(modelsMap) > 0 {
			AppConfig.LLMModels = modelsMap
			log.Println("手动加载LLMModels成功:", AppConfig.LLMModels)
		} else {
			log.Println("无法找到llm_models配置")
		}
	}

	if AppConfig.LLMSystemPrompt == "" {
		AppConfig.LLMSystemPrompt = viper.GetString("llm_system_prompt")
		log.Println("手动加载LLMSystemPrompt成功:", AppConfig.LLMSystemPrompt)
	}

	// 加载LLMCharacters
	if err := viper.UnmarshalKey("llm_characters", &AppConfig.LLMCharacters); err != nil {
		log.Printf("无法解析llm_characters配置: %v", err)
		log.Println("尝试手动加载LLMCharacters")
		charactersSlice := viper.Get("llm_characters")
		if charactersSlice != nil {
			if characters, ok := charactersSlice.([]interface{}); ok {
				AppConfig.LLMCharacters = make([]*LLMCharacter, 0, len(characters))

				for _, characterData := range characters {
					if characterMap, ok := characterData.(map[string]interface{}); ok {
						var character LLMCharacter
						if id, exists := characterMap["id"]; exists {
							character.ID = id.(string)
						}
						if name, exists := characterMap["name"]; exists {
							character.Name = name.(string)
						}
						if personality, exists := characterMap["personality"]; exists {
							character.Personality = personality.(string)
						}
						if model, exists := characterMap["model"]; exists {
							character.Model = model.(string)
						}
						if avatar, exists := characterMap["avatar"]; exists {
							character.Avatar = avatar.(string)
						}
						if customPrompt, exists := characterMap["custom_prompt"]; exists {
							character.CustomPrompt = customPrompt.(string)
						}

						AppConfig.LLMCharacters = append(AppConfig.LLMCharacters, &character)
					}
				}

				log.Printf("手动加载LLMCharacters成功: %d个角色", len(AppConfig.LLMCharacters))
			} else {
				log.Println("llm_characters格式不正确，应为数组")
			}
		} else {
			log.Println("无法找到llm_characters配置")
		}
	}

	// 加载LLMGroups
	if err := viper.UnmarshalKey("llm_groups", &AppConfig.LLMGroups); err != nil {
		log.Printf("无法解析llm_groups配置: %v", err)
		log.Println("尝试手动加载LLMGroups")
		groupsSlice := viper.Get("llm_groups")
		if groupsSlice != nil {
			if groups, ok := groupsSlice.([]interface{}); ok {
				AppConfig.LLMGroups = make([]*LLMGroup, 0, len(groups))

				for _, groupData := range groups {
					if groupMap, ok := groupData.(map[string]interface{}); ok {
						var group LLMGroup

						// 使用mapstructure或手动转换
						if id, exists := groupMap["id"]; exists {
							group.ID = id.(string)
						}
						if name, exists := groupMap["name"]; exists {
							group.Name = name.(string)
						}
						if desc, exists := groupMap["description"]; exists {
							group.Description = desc.(string)
						}
						if isGroupMode, exists := groupMap["isGroupDiscussionMode"]; exists {
							group.IsGroupDiscussionMode = isGroupMode.(bool)
						}

						// 处理members数组
						if members, exists := groupMap["members"]; exists {
							if membersArr, ok := members.([]interface{}); ok {
								group.Members = make([]string, 0, len(membersArr))
								for i, m := range membersArr {
									group.Members[i] = m.(string)
								}
							}
						}

						AppConfig.LLMGroups = append(AppConfig.LLMGroups, &group)
					}
				}

				log.Printf("手动加载LLMGroups成功: %d个群组", len(AppConfig.LLMGroups))
			} else {
				log.Println("llm_groups格式不正确，应为数组")
			}
		} else {
			log.Println("无法找到llm_groups配置")
		}
	}

	// 替换环境变量
	for provider, providerConfig := range AppConfig.LLMProviders {
		envVarName := providerConfig.APIKey
		log.Println("provider:", provider, "envVarName:", envVarName)
		if envValue := os.Getenv(envVarName); envValue != "" {
			updatedConfig := providerConfig
			updatedConfig.APIKey = envValue
			AppConfig.LLMProviders[provider] = updatedConfig
		}
	}

	// 处理模型名称中的特殊字符
	for modelName, provider := range AppConfig.LLMModels {
		if newModelName := strings.Replace(modelName, "__", ".", 1); newModelName != modelName {
			AppConfig.LLMModels[newModelName] = provider
			delete(AppConfig.LLMModels, modelName)
		}
	}
	// 处理角色model中的特殊字符
	for _, character := range AppConfig.LLMCharacters {
		if newCharacterModel := strings.Replace(character.Model, "__", ".", 1); newCharacterModel != character.Model {
			character.Model = newCharacterModel
		}
	}
}
