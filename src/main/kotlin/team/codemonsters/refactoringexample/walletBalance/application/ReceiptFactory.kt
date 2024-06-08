package team.codemonsters.refactoringexample.walletBalance.application

import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.mq.BasicJmsRequest
import java.util.UUID

interface ReceiptFactory {
    fun messageId(): String
    fun basicJmsRequest(): BasicJmsRequest<String>

    fun prepareSipHeaders(
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