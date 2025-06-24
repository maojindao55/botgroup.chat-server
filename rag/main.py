# -*- coding: utf-8 -*-
import os
from dotenv import load_dotenv
from llama_index.core import (
    VectorStoreIndex,
    SimpleDirectoryReader,
    Settings
)
from llama_index.llms.dashscope import DashScope, DashScopeGenerationModels
from llama_index.core.embeddings import BaseEmbedding
from typing import Optional, List, Any, Generator
import dashscope
from dashscope import Generation, TextEmbedding
from pydantic import Field
import numpy as np
import json
import logging

logger = logging.getLogger(__name__)

# 加载环境变量
load_dotenv()

class QwenEmbedding(BaseEmbedding):
    """千问嵌入模型包装器"""
    
    def __init__(self):
        super().__init__()
        # 设置 API key
        dashscope.api_key = os.getenv("DASHSCOPE_API_KEY")
    
    def _get_query_embedding(self, query: str) -> List[float]:
        response = TextEmbedding.call(
            model="text-embedding-v2",
            input=query
        )
        if response.status_code == 200:
            # 从 API 响应中提取嵌入向量
            embedding = response.output['embeddings'][0]['embedding']
            return embedding
        else:
            raise Exception(f"API调用失败: {response.message}")
    
    def _get_text_embedding(self, text: str) -> List[float]:
        return self._get_query_embedding(text)
    
    def _get_text_embeddings(self, texts: List[str]) -> List[List[float]]:
        response = TextEmbedding.call(
            model="text-embedding-v2",
            input=texts
        )
        if response.status_code == 200:
            # 从 API 响应中提取嵌入向量
            embeddings = [item['embedding'] for item in response.output['embeddings']]
            return embeddings
        else:
            raise Exception(f"API调用失败: {response.message}")
    
    async def _aget_query_embedding(self, query: str) -> List[float]:
        return self._get_query_embedding(query)
    
    async def _aget_text_embedding(self, text: str) -> List[float]:
        return self._get_text_embedding(text)
    
    async def _aget_text_embeddings(self, texts: List[str]) -> List[List[float]]:
        return self._get_text_embeddings(texts)

def create_dashscope_llm(model_name: str = "qwen-max", temperature: float = 0.7) -> DashScope:
    """创建 DashScope LLM 实例"""
    api_key = os.getenv("DASHSCOPE_API_KEY")
    
    # 根据模型名称选择对应的枚举值
    model_enum_map = {
        "qwen-max": DashScopeGenerationModels.QWEN_MAX,
        "qwen-plus": DashScopeGenerationModels.QWEN_PLUS,
        "qwen-turbo": DashScopeGenerationModels.QWEN_TURBO,
    }
    
    model_enum = model_enum_map.get(model_name, DashScopeGenerationModels.QWEN_MAX)
    
    return DashScope(
        model_name=model_enum,
        api_key=api_key,
        temperature=temperature
    )

def create_index(file_path: str):
    """创建文档索引"""
    # 加载文档
    documents = SimpleDirectoryReader(input_files=[file_path]).load_data()
    
    # 创建 DashScope LLM
    llm = create_dashscope_llm(temperature=0.7)
    
    # 创建千问嵌入模型
    embed_model = QwenEmbedding()
    
    # 使用新的 Settings 方式配置全局设置
    Settings.llm = llm
    Settings.embed_model = embed_model
    
    # 设置文档分块参数，增加块大小和重叠
    from llama_index.core.node_parser import SentenceSplitter
    Settings.node_parser = SentenceSplitter(
        chunk_size=1024,  # 增加块大小
        chunk_overlap=200,  # 增加重叠
        separator=" ",
    )
    
    # 创建索引
    index = VectorStoreIndex.from_documents(documents)
    return index

def main():
    # 检查 API key
    if not os.getenv("DASHSCOPE_API_KEY"):
        print("请设置 DASHSCOPE_API_KEY 环境变量")
        return

    print("正在创建文档索引...")
    index = create_index('data/sample.txt')
    
    print("\n文档索引创建完成！现在你可以开始提问了。")
    print("输入 'quit' 退出程序")
    
    while True:
        query = input("\n请输入你的问题: ")
        if query.lower() == 'quit':
            break
            
        try:
            # 创建查询引擎
            query_engine = index.as_query_engine(streaming=True)
            # 执行查询
            print("\n回答: ", end="", flush=True)
            response = query_engine.query(query)
            if hasattr(response, 'response_gen'):
                for text_chunk in response.response_gen:
                    if isinstance(text_chunk, str):
                        print(text_chunk, end="", flush=True)
                    elif hasattr(text_chunk, 'text'):
                        print(text_chunk.text, end="", flush=True)
            else:
                print(response.response, end="", flush=True)
            print()  # 打印换行
        except Exception as e:
            print("发生错误: {}".format(str(e)))

if __name__ == "__main__":
    main() 