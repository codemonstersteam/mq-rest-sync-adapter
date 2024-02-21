package team.codemonsters.refactoringexample.walletBalance

import org.springframework.messaging.Message
import org.springframework.stereotype.Service
import reactor.core.publisher.Mono
import reactor.kotlin.core.publisher.toMono
import team.codemonsters.refactoringexample.configuration.MqConfiguration
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.configuration.RestConfiguration

@Service
class Workflow(
    private val mqConfiguration: MqConfiguration,
    private val restConfiguration: RestConfiguration,
    private val brokerGateway: BrokerGateway,
    private val restClient: RestGateway,
) {
    private val mqConfig: MqServiceCfg = mqConfiguration.configs["ps"]!!

    fun process(message: Message<Any>): Mono<Result<WalletBalance>> =
        validateRequest(message)
            .pipe { sendAcceptedReceiptToQueue(it) }
            .pipe { restRequestPipe(it) }
            .pipe { sendTransactionReceiptToQueue(it) }
            .pipe { sendRestResponseToQueue(it) }

    private fun validateRequest(message: Message<Any>): Mono<Result<ValidRequest>> =
        Mono.just(
            UnvalidatedRequest.emerge(message, restConfiguration)
                .flatMap { ValidRequest.emerge(it, mqConfig.timeout) },
        )

    private fun sendAcceptedReceiptToQueue(validRequest: ValidRequest) =
        brokerGateway.sendReceipt(
            ReceiptValidationAcceptance.accepted(validRequest),
            validRequest.unvalidatedRequest.httpRequestEnvelope.operation,
            validRequest.unvalidatedRequest.incomingMessage,
            mqConfig,
        )
            .map { validRequest }
            .toMono()

    private fun restRequestPipe(validRequest: ValidRequest): Mono<Result<WalletBalance>> =
        sendRestRequest(RestServiceRequest.of(validRequest))
            .map { WalletBalance(validRequest, it) }
            .map { Result.success(it) }

    private fun sendTransactionReceiptToQueue(walletBalance: WalletBalance) =
        brokerGateway.sendReceipt(
            TransactionReceipt.receipt(walletBalance),
            walletBalance.validRequest.unvalidatedRequest.httpRequestEnvelope.operation,
            walletBalance.validRequest.unvalidatedRequest.incomingMessage,
            mqConfig,
        )
            .map { walletBalance }
            .toMono()

    private fun sendRestResponseToQueue(walletBalance: WalletBalance) =
        brokerGateway.sendResponse(
            walletBalance.httpResponseEnvelope,
            walletBalance.validRequest.unvalidatedRequest.httpRequestEnvelope.operation,
            walletBalance.validRequest.unvalidatedRequest.incomingMessage,
            mqConfig,
        )
            .map { walletBalance }
            .toMono()

    private fun sendRestRequest(request: RestServiceRequest) = restClient.sendRequest(request)
}
