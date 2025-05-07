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
}

var AppConfig Config

// LoadConfig 加载配置文件
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("server.port", "8080")

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

	// 环境变量覆盖
	if port := os.Getenv("SERVER_PORT"); port != "" {
		AppConfig.Server.Port = port
	}

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
