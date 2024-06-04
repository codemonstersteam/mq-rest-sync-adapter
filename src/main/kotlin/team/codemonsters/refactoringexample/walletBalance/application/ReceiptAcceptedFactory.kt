package team.codemonsters.refactoringexample.walletBalance.application

import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.mq.BasicJmsRequest
import team.codemonsters.refactoringexample.util.JsonUtil
import team.codemonsters.refactoringexample.walletBalance.ValidRequest
import team.codemonsters.refactoringexample.walletBalance.ensure

data class ReceiptAcceptedFactory private constructor(
    private val validRequest: ValidRequest,
    private val mqConfiguration: MqServiceCfg
) : ReceiptFactory {

    override fun messageId(): String = validRequest
        .unvalidatedRequest
        .incomingMessage
        .headers["jms_messageId"].toString()

    override fun toBasicJmsRequest(): BasicJmsRequest<String> {
        val data = ReceiptValidationAcceptance.accepted(validRequest)
        val operation = validRequest.httpRequestEnvelope.operation
        return BasicJmsRequest(
            correlationId = validRequest.httpRequestEnvelope.correlationId,
            queue = mqConfiguration.RECEIPT,
            payload = JsonUtil.toJson(data),
            headers = prepareSipHeaders(
                validRequest
                    .httpRequestEnvelope
                    .correlationId,
                operation,
                mqConfiguration
            ),
        )
    }

    companion object {
        fun emerge(
            validRequest: ValidRequest,
            mqConfig: MqServiceCfg
        ): Result<ReceiptAcceptedFactory> =
            Result.success(validRequest)
                .ensure(
                    { isCorrelationIdPresentInHeader(it, mqConfig) },
                    RuntimeException("Missing ${mqConfig.correlationID} header")
                ).ensure(
                    { isMessageIdPresentInHeader(it, mqConfig) },
                    RuntimeException("Missing ${mqConfig.correlationID} header")
                )
                .map {
                    ReceiptAcceptedFactory(validRequest, mqConfig)
                }

        private fun isMessageIdPresentInHeader(it: ValidRequest, mqConfig: MqServiceCfg) =
            it.unvalidatedRequest.incomingMessage.headers["jms_messageId"] != null

        private fun isCorrelationIdPresentInHeader(
            it: ValidRequest,
            mqConfig: MqServiceCfg
        ) = it.unvalidatedRequest.incomingMessage.headers[mqConfig.correlationID] != null
    }

}