0. 基础设施与通用能力（所有模块都会用到）

0.1 认证与授权
	•	登录/注册（邮箱/手机号/用户名）
	•	Token：JWT/Session（建议 Access + Refresh）
	•	登出、刷新 token、踢下线
	•	角色与权限：user / creator / admin
	•	RBAC / 权限中间件（路由级、资源级）
	•	账号安全：改密、找回密码、验证码（可后做）

0.2 API 基础
	•	REST（或 gRPC 内部）+ OpenAPI 文档
	•	统一错误码、请求 ID、结构化日志
	•	限流（IP/用户级）、防刷（播放/点赞/评论）
	•	幂等（上传、创建资源）
	•	审计日志（admin 操作、封禁下架等）

0.3 异步任务系统（强烈建议）
	•	任务队列：转码、抽帧、发通知、索引更新、统计聚合
	•	Worker：可水平扩展
	•	任务重试、死信队列、可观测（任务状态）

0.4 文件与媒体存储
	•	对象存储：上传原始文件、转码产物、封面、字幕
	•	CDN 分发
	•	文件鉴权：签名 URL / 私有桶临时访问
	•	断点续传/分片上传（推荐走 tus 或 S3 Multipart）

0.5 可观测性
	•	Metrics：QPS、延迟、错误率、队列堆积、转码耗时
	•	Trace（可选）：OpenTelemetry
	•	业务埋点采集：播放开始/结束、完播、停留时长（可先简化）

⸻

1. 用户系统（Users）

1.1 用户资料
	•	获取/更新用户 profile：头像、昵称、简介、地区、链接
	•	账号状态：正常/禁用/封禁（含原因与到期时间）
	•	用户设置：隐私、通知开关（可后做）

1.2 频道（Channel）
	•	channel 主页信息（对标 YouTube Channel）：横幅、简介、统计
	•	频道内容聚合：公开视频列表、短视频列表、直播回放列表、播放列表
	•	频道权限：本人编辑

⸻

2. 社交关系（Subscriptions / Follow）
	•	关注/取关频道（订阅）
	•	获取关注列表、粉丝列表
	•	订阅流 feed（按时间排序）
	•	通知触发：关注者发布新视频 / 开播（触发通知任务）

⸻

3. 视频内容（Long Video / VOD）

3.1 视频元数据与状态机
	•	创建视频记录（upload session / draft）
	•	视频字段：title、description、cover、tags、category、visibility、status
	•	可见性：public / unlisted / private / scheduled
	•	状态：draft / uploading / processing / ready / rejected / removed
	•	发布：立即发布/定时发布（定时任务）

3.2 上传与处理（Media Pipeline）
	•	获取上传凭证：预签名 URL 或上传 session
	•	断点续传/分片上传
	•	上传完成回调：触发转码任务
	•	转码（多码率）：输出 HLS/DASH（至少 HLS）
	•	抽帧：封面、预览图
	•	媒体探测：时长、分辨率、码率、fps
	•	字幕：上传/绑定字幕文件（SRT/VTT），可多语言
	•	章节（Chapter）：保存章节点（可选）

3.3 播放与鉴权
	•	获取播放信息：m3u8 master/variant 地址（可能需要签名）
	•	访问控制：private/unlisted/age/region（可后做）
	•	播放进度上报与拉取（继续观看）
	•	推荐列表：相关视频 / 作者更多 / 热门（可先简单）

3.4 互动（针对视频）
	•	点赞/取消点赞
	•	收藏/取消收藏（加入“稍后再看”或收藏夹）
	•	分享计数（可选）
	•	播放列表（playlist）创建/编辑/添加视频/排序

⸻

4. 短视频（Shorts）

短视频与长视频可以复用同一套 Video 表，只用 type=short，但 feed、播放形态、统计通常独立。

	•	短视频发布：竖屏校验（可选）
	•	短视频流：按推荐/热门/关注（支持分页、预加载友好）
	•	互动：赞/评/藏/分享、关注作者
	•	话题 hashtag（可后做）
	•	音频库/同款音乐（后做）

⸻

5. 直播（Live）

5.1 直播间管理
	•	创建直播间：title、category、cover、scheduledAt
	•	生成推流密钥与推流地址（RTMP/SRT/WebRTC 取决于方案）
	•	开播/停播状态：scheduled / live / ended
	•	在线人数/热度：实时或近实时

5.2 播放分发
	•	拉流播放地址：HLS（低延迟 LL-HLS 可后做）
	•	直播鉴权：推流鉴权、拉流鉴权（签名 URL）
	•	回放生成：录制文件 -> 转码 -> 作为 replay(VOD) 发布

5.3 直播聊天（Chat）
	•	WebSocket 聊天服务（建议独立服务/模块）
	•	消息：发送/接收、历史拉取（最近 N 条）
	•	风控：限流、敏感词、禁言、房管
	•	聊天室事件：进房、关注、礼物（礼物先做 UI/事件即可）

⸻

6. 评论系统（Comments）——视频/短视频/直播通用
	•	评论列表：按时间/热度排序
	•	发表评论、回复（两层）
	•	删除（本人/管理员）、置顶（作者/管理员）
	•	评论点赞
	•	@提及（可后做）
	•	反垃圾：频率限制、敏感词过滤、风控策略
	•	举报入口（对接举报系统）

⸻

7. 搜索与发现（Search / Discovery）

7.1 搜索
	•	视频搜索：标题、简介、标签、作者
	•	频道搜索
	•	直播搜索（直播标题、分类）
	•	排序：相关/最新/最多播放
	•	高亮与分词（后做）

实现建议：先用 PostgreSQL FTS 或 MySQL FULLTEXT；后续换 ES/Meilisearch。

7.2 发现与推荐（简化版）
	•	首页混排推荐：hot + subscriptions + recent
	•	热门榜：过去 24h/7d 播放&互动加权
	•	分类页：category feed

⸻

8. 通知系统（Notifications）
	•	通知类型：关注发布、开播提醒、评论回复、@你、审核结果
	•	站内通知：列表、已读/未读
	•	推送：邮件/短信/APP push（可后做）
	•	通知生产：事件驱动（发布/开播/回复等 -> 任务队列）

⸻

9. 审核与安全（Moderation / Trust & Safety）

9.1 内容审核
	•	视频/短视频/直播：标题、简介、封面、内容审核状态
	•	人审工作流：待审队列、审核记录
	•	驳回原因、重新提交
	•	下架/限流/封禁创作者

9.2 举报系统
	•	举报类型：内容/评论/用户/直播聊天
	•	工单状态：open / in_progress / resolved / rejected
	•	处理动作：删除、下架、封禁、警告
	•	审计日志

⸻

10. 管理后台（Admin）
	•	Admin 登录与权限
	•	用户管理：搜索、封禁/解封、角色
	•	内容管理：下架、置顶、推荐位、限流
	•	审核队列：通过/驳回
	•	举报管理：处理、备注
	•	数据看板：上传量、播放量、DAU、转码队列堆积、带宽（可后做）

⸻

11. 统计与计费（Metrics / Analytics / Monetization 可选）

11.1 统计（强烈建议至少做基础）
	•	播放量（PV/UV）
	•	观看时长、完播率、平均观看
	•	互动计数：赞/评/藏/转发
	•	直播：峰值在线、累计观看、聊天数
	•	创作者后台数据：按日聚合

11.2 变现（建议后做）
	•	广告投放与结算（复杂）
	•	会员订阅（支付、权益、内容分级）
	•	打赏/礼物（支付、风控、账务、退款）

“服务/模块”划分（单体也按这个拆目录）
	1.	auth（认证、token、角色）
	2.	users & channels
	3.	videos（long/short/replay 统一）
	4.	media（upload session、签名 URL、回调、转码任务）
	5.	feed & recommendation（首页/热门/订阅流）
	6.	comments
	7.	live（直播间、推流密钥、状态、回放）
	8.	chat（WebSocket）
	9.	search
	10.	notifications
	11.	moderation & reports
	12.	admin
	13.	analytics（计数与聚合）

MVP 优先级建议
	•	P0：auth、users/channels、videos（上传→转码→播放）、comments、likes/favorites、feed(home/trending/subscriptions)、search（基础）、studio（内容管理 API）、admin（下架/封禁/审核状态）
	•	P1：shorts feed 优化、live（开播/观看/聊天室）、回放、通知
	•	P2：审核工作流、举报、反作弊、统计聚合完善
	•	P3：商业化

⸻

思路：
	•	数据库表设计（Postgres/MySQL）核心字段
	•	API 路由清单（REST，按前端页面对齐）
	•	事件与任务队列设计（转码、通知、索引、统计）
	•	media pipeline 的最小可行方案（FFmpeg + worker + HLS + CDN）
	•	直播方案选型（RTMP+HLS / SRT / WebRTC）
