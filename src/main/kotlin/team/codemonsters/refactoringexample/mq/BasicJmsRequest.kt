package team.codemonsters.refactoringexample.mq

data class BasicJmsRequest<R>(
    val correlationId: String,
    val queue: String,
    val payload: R,
    val headers: Map<String, String>,
)