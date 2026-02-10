# TIM - Instant Messaging Backend

![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)
![Framework](https://img.shields.io/badge/Framework-Gin-0081C9?style=flat)
![Database](https://img.shields.io/badge/DB-MySQL%20%7C%20Mongo%20%7C%20Redis-green?style=flat)

TIM 是一个基于 Go 语言开发的高性能即时通讯系统后端。项目设计注重高并发处理与模块化扩展，适用于二次开发或作为 IM 基础架构。

---

## 🏗 技术栈

| 模块 | 技术选型 | 作用 |
| :--- | :--- | :--- |
| **语言** | Go (Golang) | 高并发、高性能后端支撑 |
| **Web 框架** | Gin | 提供 RESTful API 接口 |
| **消息队列** | Kafka | 异步处理消息投递，实现削峰填谷 |
| **关系型数据库** | MySQL | 存储用户信息、好友关系、群组元数据 |
| **文档数据库** | MongoDB | 存储海量聊天历史记录 |
| **缓存** | Redis | 存储在线状态、分布式 Session、验证码 |

---

## ✨ 功能特性

### 1. 核心 IM 功能
* 支持 WebSocket 长连接通信。
* 支持单聊、群聊消息转发。
* **消息加密**：可在 `ws` 包中配置服务端加解密，或由客户端自主加密，确保数据隐私。

### 2. 高可扩展性
* **文件存储**：通过重写 `pkg` 包下的接口，支持从本地存储平滑迁移至阿里云 OSS、腾讯云 COS 等。
* **验证码系统**：验证码模块接口化，支持接入不同的短信或邮件服务商。

---

## 📂 目录结构预览

```text
├── api/          # 路由与控制器层
├── db/           # 数据库连接与模型定义
├── client/       # 客户端模拟/测试脚本
├── pkg/          # 公共组件库 (OSS、加密、验证码接口)
├── config.yaml   # 配置文件
└── main.go       # 项目入口
