#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
简化版微信登录测试脚本
不依赖任何外部库，只使用 Python 标准库
"""

import json
import time
import uuid
import subprocess
import urllib.request
import urllib.parse


def redis_cmd(*args):
    """执行 Redis 命令"""
    try:
        cmd = ['redis-cli', '-a', 'redis123'] + list(args)
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        return result.stdout.strip()
    except (subprocess.CalledProcessError, FileNotFoundError) as e:
        print(f"❌ Redis 命令失败: {e}")
        return None


def test_redis():
    """测试 Redis 连接"""
    result = redis_cmd('ping')
    if result == 'PONG':
        print("✅ Redis 连接成功")
        return True
    else:
        print("❌ Redis 连接失败")
        return False


def create_session():
    """创建登录会话"""
    if not test_redis():
        return None
    
    session_id = str(uuid.uuid4())
    qr_scene = "login_1756336918_33da3ae967915207"
    now = int(time.time())
    expires_at = now + 600
    
    session_data = {
        "session_id": session_id,
        "qr_scene": qr_scene,
        "status": "pending",
        "user_id": 0,
        "openid": "",
        "created_at": now,
        "expires_at": expires_at
    }
    
    redis_key = f"wechat_login:session:{session_id}"
    session_json = json.dumps(session_data)
    
    result = redis_cmd('SETEX', redis_key, '600', session_json)
    
    if result == 'OK':
        print("=== 第1步：创建登录会话 ===")
        print(f"✅ Session ID: {session_id}")
        print(f"✅ QR Scene: {qr_scene}")
        print(f"✅ 已保存到 Redis")
        print("=" * 50)
        return session_id, qr_scene
    else:
        print("❌ 创建会话失败")
        return None, None


def send_callback(qr_scene):
    """发送微信回调请求"""
    print("=== 第2步：发送微信回调请求 ===")
    
    url = "http://localhost:8082/api/auth/wechat/callback"
    
    # 查询参数
    params = {
        'signature': '7c95059ee4a89926fcb2321af75f0f3ab22da603',
        'timestamp': '1756336927',
        'nonce': '1386465447',
        'openid': 'ofOmgvncLMs1b0CKdrSGnr6WmqO0'
    }
    
    # 构建完整 URL
    query_string = urllib.parse.urlencode(params)
    full_url = f"{url}?{query_string}"
    
    # XML 消息体
    xml_data = f"""<xml>
<ToUserName><![CDATA[gh_f7dbc5fd2d54]]></ToUserName>
<FromUserName><![CDATA[ofOmgvncLMs1b0CKdrSGnr6WmqO0]]></FromUserName>
<CreateTime>1756336927</CreateTime>
<MsgType><![CDATA[event]]></MsgType>
<Event><![CDATA[SCAN]]></Event>
<EventKey><![CDATA[{qr_scene}]]></EventKey>
<Ticket><![CDATA[gQHN8DwAAAAAAAAAAS5odHRwOi8vd2VpeGluLnFxLmNvbS9xLzAyTFg5aVEwWUZhdy0xRkxsTE5FY2UAAgQXk69oAwRYAgAA]]></Ticket>
</xml>"""
    
    try:
        # 创建请求
        req = urllib.request.Request(
            full_url,
            data=xml_data.encode('utf-8'),
            headers={
                'Content-Type': 'application/xml; charset=utf-8',
                'User-Agent': 'Mozilla/4.0'
            },
            method='POST'
        )
        
        # 发送请求
        with urllib.request.urlopen(req, timeout=10) as response:
            response_text = response.read().decode('utf-8')
            status_code = response.getcode()
            
            print(f"📥 响应状态码: {status_code}")
            print(f"📥 响应内容: {response_text}")
            
            if status_code == 200:
                print("✅ 微信回调请求成功")
                return True
            else:
                print("❌ 微信回调请求失败")
                return False
    
    except Exception as e:
        print(f"❌ 请求失败: {e}")
        return False


def check_session(session_id):
    """检查会话状态"""
    print("=== 第3步：检查会话状态 ===")
    
    redis_key = f"wechat_login:session:{session_id}"
    data = redis_cmd('GET', redis_key)
    
    if not data:
        print("❌ 会话不存在或已过期")
        return False
    
    try:
        session = json.loads(data)
        
        print(f"📋 Session ID: {session.get('session_id')}")
        print(f"📋 QR Scene: {session.get('qr_scene')}")
        print(f"📋 Status: {session.get('status')}")
        print(f"📋 User ID: {session.get('user_id', 0)}")
        print(f"📋 OpenID: {session.get('openid', '')}")
        
        if session.get('status') == 'success':
            print("✅ 登录成功！会话状态已更新")
            return True
        elif session.get('status') == 'pending':
            print("⏳ 会话仍在等待状态")
            return False
        else:
            print(f"⚠️  会话状态: {session.get('status')}")
            return False
    
    except Exception as e:
        print(f"❌ 检查会话状态失败: {e}")
        return False


def main():
    """主测试流程"""
    print("🚀 开始微信登录完整流程测试（简化版）")
    print("=" * 60)
    
    # 步骤1：创建会话
    session_id, qr_scene = create_session()
    if not session_id:
        print("❌ 测试终止：无法创建会话")
        return
    
    # 步骤2：发送回调
    if not send_callback(qr_scene):
        print("❌ 测试终止：回调请求失败")
        return
    
    # 等待服务器处理
    time.sleep(1)
    
    # 步骤3：检查结果
    success = check_session(session_id)
    
    print("=" * 60)
    if success:
        print("🎉 测试成功！微信登录流程正常工作")
    else:
        print("⚠️  测试部分成功，但会话状态未更新")
        print("   可能的原因:")
        print("   1. 签名验证失败")
        print("   2. 服务器配置问题")
        print("   3. 数据库连接问题")
    
    print("\n🏁 测试完成")


if __name__ == "__main__":
    main()