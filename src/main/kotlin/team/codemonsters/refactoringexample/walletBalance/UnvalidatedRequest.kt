package team.codemonsters.refactoringexample.walletBalance

import org.springframework.messaging.Message
import team.codemonsters.refactoringexample.configuration.RestConfiguration

data class UnvalidatedRequest private constructor(
    val incomingMessage: Message<Any>,
    val restConfiguration: RestConfiguration,
    val httpRequestEnvelope: HttpRequestEnvelope,
) {
    companion object {
        fun emerge(
            incomingMessage: Message<Any>,
            restConfiguration: RestConfiguration,
        ): Result<UnvalidatedRequest> =
            HttpRequestEnvelope.fromJsonString(incomingMessage.payload.toString())
                .map {
                    UnvalidatedRequest(
                        incomingMessage,
                        restConfiguration,
                        it,
                    )
                }
    }
}
