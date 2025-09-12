module github.com/anasamu/microservices-library-go/ai

go 1.21

require (
	github.com/anasamu/microservices-library-go/ai/providers/anthropic v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/ai/providers/deepseek v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/ai/providers/google v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/ai/providers/openai v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/ai/providers/xai v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/ai/types v0.0.0
)

replace github.com/anasamu/microservices-library-go/ai/providers/anthropic => ./providers/anthropic

replace github.com/anasamu/microservices-library-go/ai/providers/deepseek => ./providers/deepseek

replace github.com/anasamu/microservices-library-go/ai/providers/google => ./providers/google

replace github.com/anasamu/microservices-library-go/ai/providers/openai => ./providers/openai

replace github.com/anasamu/microservices-library-go/ai/providers/xai => ./providers/xai

replace github.com/anasamu/microservices-library-go/ai/types => ./types
