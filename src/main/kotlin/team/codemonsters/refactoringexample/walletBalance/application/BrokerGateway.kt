package team.codemonsters.refactoringexample.walletBalance.application

import org.slf4j.LoggerFactory
import org.springframework.messaging.Message
import org.springframework.stereotype.Component
import reactor.core.publisher.Mono
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.mq.BasicJmsRequest
import team.codemonsters.refactoringexample.service.MqPublisher
import team.codemonsters.refactoringexample.util.JsonUtil
import java.util.UUID

@Component
class BrokerGateway(
    private val mqPublisher: MqPublisher,
) {
    private val log = LoggerFactory.getLogger(BrokerGateway::class.java)!!

    fun sendReceipt(
        receipt: ReceiptFactory,
    ): Mono<Result<BasicJmsRequest<String>>> =
        Mono.just(
            runCatching {
                mqPublisher.sendToMq(receipt.basicJmsRequest(), receipt.messageId())
                receipt.basicJmsRequest()
            }
                .onFailure { log.error("Error sendReceipt ${it.message}") },
        )

}
