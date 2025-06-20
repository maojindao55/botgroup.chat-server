from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
from typing import Optional
from contextlib import asynccontextmanager
import uvicorn
from main import create_index, create_dashscope_llm
import json
import logging
import asyncio
import os
from dotenv import load_dotenv
from llama_index.core import Settings

# 加载环境变量
load_dotenv()

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
logging.getLogger("dashscope").setLevel(logging.WARNING)

# 全局变量存储索引
index = None

class QueryRequest(BaseModel):
    query: str
    temperature: Optional[float] = 0.7
    model_name: Optional[str] = "qwen-max"
    custom_prompt: Optional[str] = None  # 允许用户自定义prompt

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时执行
    global index
    try:
        # 检查 API key
        if not os.getenv("DASHSCOPE_API_KEY"):
            raise Exception("请设置 DASHSCOPE_API_KEY 环境变量")
        
        logger.info("开始初始化索引...")
        index = create_index()
        logger.info("索引初始化完成")
        
    except Exception as e:
        logger.error(f"初始化失败: {str(e)}", exc_info=True)
        raise Exception(f"初始化失败: {str(e)}")
    
    yield
    
    # 关闭时执行（如果需要清理资源）
    logger.info("应用关闭")

app = FastAPI(
    title="文档问答API", 
    description="基于千问大模型的文档问答服务",
    lifespan=lifespan
)

async def stream_generator(request: QueryRequest):
    try:
        logger.debug("开始执行RAG查询...")
        
        # 创建一个临时的 DashScope 实例，使用请求中的参数
        temp_llm = create_dashscope_llm(
            model_name=request.model_name, 
            temperature=request.temperature
        )
        
        # 临时设置全局 LLM（用于查询引擎）
        original_llm = Settings.llm
        Settings.llm = temp_llm
        
        try:
            # 自定义 prompt 模板
            from llama_index.core.prompts import PromptTemplate
            
            # 根据请求决定使用哪个 prompt
            if request.custom_prompt:
                # 使用用户提供的自定义 prompt
                # 确保包含必要的占位符
                if "{context_str}" not in request.custom_prompt or "{query_str}" not in request.custom_prompt:
                    # 如果用户没有包含必要占位符，就在后面追加
                    prompt_text = request.custom_prompt + "\n\n上下文信息:\n{context_str}\n\n问题: {query_str}\n回答: "
                else:
                    prompt_text = request.custom_prompt
                custom_qa_prompt = PromptTemplate(prompt_text)
            else:
                # 使用默认的 prompt
                custom_qa_prompt = PromptTemplate(
                    "你是一个专业的文档问答助手。请基于以下提供的上下文信息来回答用户的问题。\n"
                    "如果上下文中没有相关信息，请诚实地说你不知道，不要编造答案。\n"
                    "请用中文回答，并尽量提供详细和准确的信息。\n\n"
                    "上下文信息:\n"
                    "{context_str}\n\n"
                    "问题: {query_str}\n"
                    "回答: "
                )
            logger.debug(f"custom_qa_prompt: {custom_qa_prompt}")
            # 使用 LlamaIndex 的查询引擎进行流式查询
            query_engine = index.as_query_engine(
                streaming=True,
                similarity_top_k=15,  # 增加检索的文档数量
                response_mode="tree_summarize",  # 使用更好的响应模式
                text_qa_template=custom_qa_prompt  # 自定义问答prompt
            )
            
            logger.debug("开始处理来自查询引擎的流...")
            response = query_engine.query(request.query)
            
            # 处理流式响应
            if hasattr(response, 'response_gen') and response.response_gen:
                # 如果有 response_gen 属性，说明是流式响应
                for chunk in response.response_gen:
                    if chunk:
                        # 处理不同类型的响应块
                        if isinstance(chunk, str):
                            text_chunk = chunk
                        elif hasattr(chunk, 'delta') and chunk.delta:
                            text_chunk = chunk.delta
                        elif hasattr(chunk, 'text') and chunk.text:
                            text_chunk = chunk.text
                        else:
                            text_chunk = str(chunk)
                        
                        if text_chunk:
                            yield f"data: {json.dumps({'content': text_chunk}, ensure_ascii=False)}\n\n"
                            await asyncio.sleep(0.01)
            else:
                # 如果不是流式响应，直接返回完整响应
                response_text = str(response.response) if hasattr(response, 'response') else str(response)
                yield f"data: {json.dumps({'content': response_text}, ensure_ascii=False)}\n\n"
                
        finally:
            # 恢复原始的 LLM 设置
            Settings.llm = original_llm

    except Exception as e:
        logger.error(f"处理RAG查询时发生错误: {str(e)}", exc_info=True)
        yield f"data: {json.dumps({'error': str(e)}, ensure_ascii=False)}\n\n"

@app.post("/rag/query")
async def query_documents(request: QueryRequest):
    if index is None:
        raise HTTPException(status_code=500, detail="索引未初始化")
    
    logger.info(f"收到查询请求: {request.query}, 模型: {request.model_name}, 温度: {request.temperature}")
    
    return StreamingResponse(
        stream_generator(request),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
            "Access-Control-Allow-Headers": "*",
        }
    )

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8070) 