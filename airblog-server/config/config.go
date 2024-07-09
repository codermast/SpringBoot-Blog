package config

import (
	"codermast.com/airblog/models"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
)

var (
	config     *models.Config
	onceConfig sync.Once
)

func AirBlogSystemConfig() {
	// 仅加载一次
	onceConfig.Do(func() {
		// 1. 加载配置文件
		file, err := ioutil.ReadFile("config.yaml")
		if err != nil {
			log.Printf("Error reading config file: %v", err)
			return
		}

		// 2. 解析配置文件到全局变量 config
		err = yaml.Unmarshal(file, &config)
		if err != nil {
			log.Printf("Error parsing config file: %v", err)
			return
		}
	})
}

// GetDatabaseConfig 获取数据库配置信息
func GetDatabaseConfig() *models.Database {
	AirBlogSystemConfig() // 确保配置加载完成
	return &config.Database
}

// GetJWTConfig 获取 JWT 配置信息
func GetJWTConfig() *models.JWT {
	AirBlogSystemConfig() // 确保配置加载完成
	return &config.JWT
}
