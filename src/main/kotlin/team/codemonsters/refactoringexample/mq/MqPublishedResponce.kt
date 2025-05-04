package team.codemonsters.refactoringexample.mq

data class MqPublishedResponce<T>(
    val status: String,
    val correlationId: String,
    val jmsRequest: BasicJmsRequest<T>
)