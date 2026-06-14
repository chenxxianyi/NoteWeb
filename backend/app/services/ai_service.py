"""AI mock service — returns hard-coded responses."""


class AiService:
    async def get_summary(self, document_id: int) -> str:
        return (
            "本文档主要介绍了相关主题的核心概念和应用场景。"
            "内容涵盖了基础知识、关键技术和实践方法，"
            "适合初学者和有一定经验的读者参考学习。"
        )

    async def explain(self, text: str) -> str:
        return (
            f"这段话的核心意思是：{text[:50]}…… "
            "它主要阐述了该领域的一个重要概念，"
            "强调了理论与实践相结合的重要性。"
        )

    async def translate(self, text: str, target_lang: str = "zh") -> str:
        if target_lang == "zh":
            return f"[翻译] {text}"
        return f"[Translation] {text}"

    async def chat(self, document_id: int, question: str) -> str:
        return (
            f"关于您的问题「{question}」，"
            "根据文档内容分析，这涉及到多个方面的因素。"
            "建议您结合文档中的实例和最佳实践来深入理解。"
        )
