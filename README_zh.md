# Pomodoro

一个使用 Go + Fyne 开发的桌面番茄钟应用，支持本地 SQLite 数据存储。

![效果图](/img/img.png)

[English](/README.md)

## 功能

- 番茄钟倒计时
- 工作 / 短休息 / 长休息切换
- 开始、暂停、重置、跳过
- 今日专注统计
- 历史记录查看与批量删除
- 系统通知提醒
- 本地设置持久化

## 技术栈

- Go
- Fyne
- SQLite

## 运行

```bash
go run ./cmd/pomodoro/main.go
```

## 构建

```bash
# 直接指定路径编译
go build -o pomodoro ./cmd/pomodoro/main.go
# 使用make
make build
# 同时编译Windows和Linux
make build-all
```

## 数据存储

应用设置和历史记录默认保存在本地 SQLite 数据库中。
