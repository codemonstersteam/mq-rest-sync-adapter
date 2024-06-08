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

    override fun basicJmsRequest(): BasicJmsRequest<String> {
        val data = ReceiptValidationAcceptance.accepted(validRequest)
        val operation = validRequest.httpRequestEnvelope.operation
        val correlationId = validRequest.unvalidatedRequest.incomingMessage.headers["CorrelationID"].toString()
        return BasicJmsRequest(
            correlationId = correlationId,
            queue = mqConfiguration.RECEIPT,
            payload = JsonUtil.toJson(data),
            headers = prepareSipHeaders(
                correlationId,
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
                    { isMessageIdPresentInHeader(it) },
                    RuntimeException("Missing jms_messageId header")
                )
                .map {
                    ReceiptAcceptedFactory(validRequest, mqConfig)
                }

        private fun isMessageIdPresentInHeader(it: ValidRequest,) =
            it.unvalidatedRequest.incomingMessage.headers["jms_messageId"] != null

        private fun isCorrelationIdPresentInHeader(
            it: ValidRequest,
            mqConfig: MqServiceCfg
        ) = it.unvalidatedRequest.incomingMessage.headers[mqConfig.correlationID] != null
    }

}