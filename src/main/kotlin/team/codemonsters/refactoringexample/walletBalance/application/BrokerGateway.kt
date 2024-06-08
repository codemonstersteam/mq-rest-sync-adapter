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
                val jmsRequest = receipt.basicJmsRequest()
                mqPublisher.sendToMq(jmsRequest, receipt.messageId())
                jmsRequest
            }
                .onFailure { log.error("Error sendReceipt ${it.message}") },
        )

    fun sendResponse(
        data: Any,
        operation: String,
        incomingMessage: Message<Any>,
        mqConfiguration: MqServiceCfg,
    ): Mono<Result<BasicJmsRequest<String>>> =
        Mono.just(
            runCatching {
                val correlationId = incomingMessage.headers[mqConfiguration.correlationID] as String
                BasicJmsRequest(
                    correlationId = correlationId,
                    queue = mqConfiguration.RES,
                    payload = JsonUtil.toJson(data),
                    headers = prepareSipHeaders(correlationId, operation, mqConfiguration),
                )
            }.map {
                mqPublisher.sendToMq(it, incomingMessage.headers["jms_messageId"] as String?)
                return@map it
            }
                .onFailure { log.error("Error sendResponse ${it.message}") },
        )

    private fun prepareSipHeaders(
        correlationId: String,
        url: String,
        mqConfig: MqServiceCfg,
    ): Map<String, String> {
        val headers =
            mutableMapOf(
                mqConfig.correlationID to correlationId,
                "X_Method" to url,
                "RequestID" to UUID.randomUUID().toString(),
                mqConfig.createDateTime to mqConfig.getFormattedCreateDateTime(),
            )
        headers.putAll(mqConfig.values)
        return headers
    }
}
