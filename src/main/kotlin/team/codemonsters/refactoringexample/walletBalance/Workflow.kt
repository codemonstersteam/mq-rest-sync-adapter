package team.codemonsters.refactoringexample.walletBalance

import org.slf4j.LoggerFactory
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
    private val log = LoggerFactory.getLogger(Workflow::class.java)!!
    private val mqConfig: MqServiceCfg = mqConfiguration.configs["ps"]!!
    private var flowIsNotLogged: Boolean = true

    fun process(message: Message<Any>): Mono<Result<WalletBalance>> =
        validateRequest(message)
            .doOnNext { logWalletBalanceFlow() }
            .pipe { sendAcceptedReceiptToQueue(it) }
            .pipe { getBalanceByRest(it) }
            .pipe { sendRestResponseToQueue(it) }
            .pipe { sendTransactionReceiptToQueue(it) }
            .flatMap { logError(it) }

    private fun logWalletBalanceFlow() {
        if (flowIsNotLogged) {
            log.info("wallet-balance workflow is enabled")
            flowIsNotLogged = false
        }
    }

    private fun validateRequest(message: Message<Any>): Mono<Result<ValidRequest>> =
        Mono.just(
            ValidRequest.emerge(
                UnvalidatedRequest(message, restConfiguration, mqConfig.timeout),
            )
        )

    private fun sendAcceptedReceiptToQueue(validRequest: ValidRequest): Mono<Result<ValidRequest>> =
        brokerGateway.sendReceipt(
            ReceiptValidationAcceptance.accepted(validRequest),
            validRequest.httpRequestEnvelope.operation,
            validRequest.unvalidatedRequest.incomingMessage,
            mqConfig,
        )
            .map { it.map { validRequest } }

    private fun getBalanceByRest(validRequest: ValidRequest): Mono<Result<WalletBalance>> =
        sendRestRequest(RestServiceRequest.of(validRequest))
            .map { WalletBalance(validRequest, it) }
            .map { Result.success(it) }

    private fun sendTransactionReceiptToQueue(walletBalance: WalletBalance): Mono<Result<WalletBalance>> =
        brokerGateway.sendReceipt(
            TransactionReceipt.receipt(walletBalance),
            walletBalance.validRequest.httpRequestEnvelope.operation,
            walletBalance.validRequest.unvalidatedRequest.incomingMessage,
            mqConfig,
        )
            .map { it.map { walletBalance } }

    private fun sendRestResponseToQueue(walletBalance: WalletBalance): Mono<Result<WalletBalance>> =
        brokerGateway.sendResponse(
            walletBalance.httpResponseEnvelope,
            walletBalance.validRequest.httpRequestEnvelope.operation,
            walletBalance.validRequest.unvalidatedRequest.incomingMessage,
            mqConfig,
        )
            .mapResult { Result.success(walletBalance) }

    private fun sendRestRequest(request: RestServiceRequest) = restClient.sendRequest(request)

    private fun logError(walletBalance: Result<WalletBalance>): Mono<Result<WalletBalance>> =
        walletBalance.onFailure {
            log.error("Error in wallet balance workflow ${walletBalance.exceptionOrNull()!!.message}")
        }.toMono()
}
