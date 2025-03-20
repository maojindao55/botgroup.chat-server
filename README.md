# BotGroup.Chat服务器版

## 项目背景
BotGroup.Chat 是一个基于 React 的多人 AI 聊天应用，支持多个 AI 角色同时参与对话，提供类似群聊的交互体验。

> 🔗 原项目地址：[botgroup.chat](https://github.com/maojindao55/botgroup.chat)


## 为什么要做botgroup.chat服务器版？
原项目仅支持 Cloudflare Pages 部署，这导致：
- 服务只能部署在海外
- 存在访问限制
- 部署选项单一

本项目通过 Docker 化改造，解决了以上问题，让您能够：
- 体验botgroup.chat所有特性
- 使用Docker 一键部署
- 可部署在任意服务器或者本地电脑
- 获得更好的访问速度

## 部署和安装
1. 克隆仓库
```bash
git clone https://github.com/maojindao55/botgroup.chat-server
```

2. 安装依赖
- 安装 docker 
- 安装 docker-compose
- [如何安装? 请访问docker官网](https://www.docker.com/)


3. 更新模型配置`.env.api`
```bash
mv .env.api.example .env.api (或直接更改后缀)
# 打开配置文件.env.api, 请到各个模型厂商自助申请apikey并更新以下配置
DASHSCOPE_API_KEY=your_dashscope_api_key_here
HUNYUAN_API_KEY=your_hunyuan_api_key_here
ARK_API_KEY=your_ark_api_key_here
GLM_API_KEY=your_glm_api_key_here
DEEPSEEK_API_KEY=your_deepseek_api_key_here
KIMI_API_KEY=your_kimi_api_key_here
BAIDU_API_KEY=your_baidu_api_key_here
HUNYUAN_API_KEY1=your_hunyuan_api_key1_here 
```
APIKEY|对应角色|服务商|申请地址|
|------|-----|-------|------|
|DASHSCOPE_API_KEY|千问|阿里云|https://www.aliyun.com/product/bailian|
|HUNYUAN_API_KEY|元宝|腾讯云|[新户注册免费200万tokens额度](https://cloud.tencent.com/product/hunyuan)|
|ARK_API_KEY|豆包|火山引擎|[火山引擎大模型新客使用豆包大模型及 DeepSeek R1模型各可享 10 亿 tokens/模型的5折优惠 ，5个模型总计 50 亿 tokens](https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?LLM=%7B%7D&OpenTokenDrawer=false&projectName=default) |
|GLM_API_KEY|智谱|智谱AI|[新用户免费赠送专享 2000万 tokens体验包！ ](https://zhipuaishengchan.datasink.sensorsdata.cn/t/9z)|
|DEEPSEEK_API_KEY|DeepSeek|DeepSeek|https://platform.deepseek.com|
|KIMI_API_KEY|Kimi|Moonshot AI|https://platform.moonshot.cn|
|BAIDU_API_KEY|文小言|百度千帆|https://cloud.baidu.com/campaign/qianfan|

4. 一键启动
```bash
#进入根目录执行命令：
docker-compose up -d

#默认访问地址 
http://localhost:8082

#可根据自己需求 修改 docker-compopse.yaml中端口地址
...
ports:
  - "8082:80"
...

```


4. 群聊和成员配置说明`config.yaml`(非必须)
```yaml
# 打开配置文件 src/config/config.yaml

llm_models:
    qwen-plus: "aliyun"
    qwen-turbo: "aliyun"
    ...

llm_characters:
  #第0个角色为调度器，建议不要删除。
  - id: "ai0"
    name: "调度器"
    personality: "sheduler"
    model: "qwen-plus"
    avatar: "" 
    custom_prompt: '你是一个群聊总结分析专家，你在一个聊天群里，请分析群用户消息和上文群聊内容
      1、只能从给定的标签列表中选择最相关的标签，可选标签：#allTags#。
      2、请只返回标签列表，用逗号分隔，不要有其他解释, 不要有任何前缀。
      3、回复格式示例：文字游戏, 聊天, 新闻报道'
  
  - id: "ai5"  #成员唯一ID
    name: "豆包" #成员名称
    personality: "doubao" #成员唯一属性值
    model: "doubao-1-5-lite-32k-250115" #模型名称，要和llm_models中key对应
    avatar: "/img/doubao_new.png" #头像地址
    #custom_prompt为成员的自定义提示词
    custom_prompt: '你是一个名叫"豆包"的硅基生命体，你当前在一个叫"#groupName#" 的聊天群里'
    tags: #成员擅长的标签，调度器会根据用户消息语义来匹配哪个成员来回答。
      - "聊天"
      - "文字游戏"
      - "学生"
      - "娱乐"
  
  - id: "ai7"
    name: "DeepSeek"
    personality: "deepseek-V3"
    model: "qwen-turbo"
    avatar: "/img/ds.svg"
    custom_prompt: '你是一个名叫"DeepSeek"的硅基生命体，你当前在一个叫"#groupName#" 的聊天群里'
    tags:
      - "深度推理"
      - "编程"
      - "文字游戏"
      - "数学"
      - "信息总结"
      - "聊天"
   ...


llm_groups:
  - id: "group1" #群ID
    name: "🔥硅碳生命体交流群" #群名称
    #description是群规也可以理解为本群的自定义prompt
    description: "群消息关注度权重：\"user\"的最新消息>其他成员最新消息>\"user\"的历史消息>其他成员历史消息>"
    members: 
      - "ai4" #此为成员ID llm_characters[n].id要对应
      - "ai5"
      - "ai6"
    isGroupDiscussionMode: true #是否默认打开全员讨论模式
 ...
 ...

```


## 贡献指南
欢迎提交 Pull Request 或提出 Issue。
当然也可以加共建QQ群交流：922322461（群号）

## 跪谢赞助商ORZ
此项目开源上线以来，用户猛增tokens消耗每日近千万，因此接受了国内多个基座模型厂商给予的tokens的赞助，作为开发者由衷地感谢国产AI模型服务商雪中送炭，雨中送伞！

## Tokens 赞助情况

|品牌logo  | AI服务商 | 赞助Tokens 额度 |新客注册apikey活动|
|---------|----------|------------|-------|
|![智谱AI](https://raw.githubusercontent.com/maojindao55/botgroup.chat/refs/heads/main/public/img/bigmodel.png)| 智谱AI | 5.5亿 | [新用户免费赠送专享 2000万 tokens体验包！ ](https://zhipuaishengchan.datasink.sensorsdata.cn/t/9z)|
|![火山引擎](https://portal.volccdn.com/obj/volcfe/logo/appbar_logo_dark.2.svg)| 字节跳动火山引擎 | 5亿 | 1. [火山引擎大模型新客使用豆包大模型及 DeepSeek R1模型各可享 10 亿 tokens/模型的5折优惠 ，5个模型总计 50 亿 tokens](https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement?LLM=%7B%7D&OpenTokenDrawer=false&projectName=default) <br> <br> 2. [应用实验室助力企业快速构建大模型应用，开源易集成，访问Github获取应用源代码](https://github.com/volcengine/ai-app-lab/tree/main)|
|![腾讯云](https://cloudcache.tencent-cloud.com/qcloud/portal/kit/images/slice/logo.23996906.svg)| 腾讯混元AI模型 | 1亿 |[新户注册免费200万tokens额度](https://cloud.tencent.com/product/hunyuan)|
|![monica](https://files.monica.cn/assets/botgroup/monica.png)| Monica团队 | 其他未认领模型所有tokens |[用monica中文版免费和 DeepSeek V3 & R1 对话](https://monica.cn/)|

## 许可证

[MIT License](LICENSE)