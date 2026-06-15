package service

type AIService struct{}

func NewAIService() *AIService {
	return &AIService{}
}

type SummaryResponse struct {
	Summary string `json:"summary"`
}

type ExplainResponse struct {
	Explanation string `json:"explanation"`
}

type TranslateResponse struct {
	Translation string `json:"translation"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
}

func (s *AIService) GetSummary(docID uint) SummaryResponse {
	return SummaryResponse{
		Summary: "本文档主要介绍了相关主题的核心概念和应用场景。内容涵盖了基础知识、关键技术和实践方法，适合初学者和有一定经验的读者参考学习。",
	}
}

func (s *AIService) Explain(text string) ExplainResponse {
	return ExplainResponse{
		Explanation: "这段话的核心意思是：" + text + "——它主要阐述了该领域的一个重要概念，强调了理论与实践相结合的重要性。",
	}
}

func (s *AIService) Translate(text, targetLang string) TranslateResponse {
	return TranslateResponse{
		Translation: "[翻译] " + text,
	}
}

func (s *AIService) Chat(docID uint, question string) ChatResponse {
	return ChatResponse{
		Answer: "关于您的问题「" + question + "」，根据文档内容分析，这涉及到多个方面的因素。建议您结合文档中的实例和最佳实践来深入理解。",
	}
}
