package team.codemonsters.refactoringexample.walletBalance.application

import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.mq.BasicJmsRequest
import team.codemonsters.refactoringexample.util.JsonUtil
import team.codemonsters.refactoringexample.walletBalance.WalletBalance

data class TransactionReceiptFactory(
    val walletBalance: WalletBalance,
    val mqConfiguration: MqServiceCfg
) : ReceiptFactory {

    override fun messageId(): String = walletBalance
        .validRequest
        .unvalidatedRequest
        .incomingMessage
        .headers["jms_messageId"].toString()

    override fun basicJmsRequest(): BasicJmsRequest<String> {
        val data = TransactionReceipt.receipt(walletBalance)
        val operation = walletBalance.validRequest.httpRequestEnvelope.operation
        return BasicJmsRequest(
            correlationId = walletBalance.validRequest.httpRequestEnvelope.correlationId,
            queue = mqConfiguration.RECEIPT,
            payload = JsonUtil.toJson(data),
            headers = prepareSipHeaders(
                walletBalance.validRequest.httpRequestEnvelope.correlationId,
                operation,
                mqConfiguration
            ),
        )
    }

}