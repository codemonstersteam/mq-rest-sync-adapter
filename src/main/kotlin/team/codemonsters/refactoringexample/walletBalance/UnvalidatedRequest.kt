package team.codemonsters.refactoringexample.walletBalance

import org.springframework.messaging.Message
import team.codemonsters.refactoringexample.configuration.RestConfiguration

data class UnvalidatedRequest (
    val incomingMessage: Message<Any>,
    val restConfiguration: RestConfiguration,
    val requestLifetime: Long
) {
    val payload = incomingMessage.payload
}