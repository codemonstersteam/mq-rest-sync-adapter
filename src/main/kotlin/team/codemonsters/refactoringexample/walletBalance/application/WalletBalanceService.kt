package team.codemonsters.refactoringexample.walletBalance.application

import org.slf4j.LoggerFactory
import org.springframework.messaging.Message
import org.springframework.stereotype.Service
import reactor.core.publisher.Mono
import reactor.core.scheduler.Schedulers
import reactor.kotlin.core.publisher.toMono
import team.codemonsters.refactoringexample.configuration.MqConfiguration
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.walletBalance.UnvalidatedRequest
import team.codemonsters.refactoringexample.walletBalance.ValidRequest
import team.codemonsters.refactoringexample.walletBalance.WalletBalance
import team.codemonsters.refactoringexample.walletBalance.fold
import team.codemonsters.refactoringexample.walletBalance.infrastructure.RestServiceRequest
import team.codemonsters.refactoringexample.walletBalance.infrastructure.RestServiceResponse
import team.codemonsters.refactoringexample.walletBalance.pipe

@Service
class WalletBalanceService(
    private val mqConfiguration: MqConfiguration,
    private val restConfiguration: RestConfiguration,
    private val brokerGateway: BrokerGateway,
    private val restClient: RestGateway,
) {
    private val log = LoggerFactory.getLogger(WalletBalanceService::class.java)!!
    private val mqConfig: MqServiceCfg = mqConfiguration.configs["ps"]!!

    fun process(message: Message<Any>): Mono<Result<WalletBalance>> =
        validateRequest(message)
            .pipe { sendAcceptedReceipt(it) }
            .pipe { getBalance(it) }
            .pipe { sendRestResponse(it) }
            .pipe { sendTransactionReceipt(it) }
            .flatMap { logError(it) }

    private fun validateRequest(message: Message<Any>): Mono<Result<ValidRequest>> =
        Mono.just(
            ValidRequest.emerge(
                UnvalidatedRequest(message, restConfiguration, mqConfig.timeout),
            ),
        )

    private fun sendAcceptedReceipt(validRequest: ValidRequest): Mono<Result<ValidRequest>> =
        ReceiptAcceptedFactory
            .emerge(validRequest, mqConfig)
            .toMono()
            .pipe { brokerGateway.sendReceipt(it) }
            .subscribeOn(Schedulers.boundedElastic())
            .fold { validRequest }

    private fun getBalance(validRequest: ValidRequest): Mono<Result<WalletBalance>> =
        sendRestRequest(RestServiceRequest.of(validRequest))
            .fold { WalletBalance(validRequest, it) }

    private fun sendTransactionReceipt(walletBalance: WalletBalance): Mono<Result<WalletBalance>> =
        brokerGateway
            .sendReceipt(TransactionReceiptFactory(walletBalance, mqConfig))
            .map { it.map { walletBalance } }

    private fun sendRestResponse(walletBalance: WalletBalance): Mono<Result<WalletBalance>> =
        brokerGateway
            .sendReceipt(ResponseReceiptFactory(walletBalance, mqConfig))
            .subscribeOn(Schedulers.boundedElastic())
            .fold { walletBalance }

    private fun sendRestRequest(request: RestServiceRequest): Mono<Result<RestServiceResponse>> =
        restClient
            .sendRequest(request)
            .subscribeOn(Schedulers.boundedElastic())

    private fun logError(walletBalance: Result<WalletBalance>): Mono<Result<WalletBalance>> =
        walletBalance
            .onFailure {
                log.error("Error in wallet balance workflow ${walletBalance.exceptionOrNull()!!.message}")
            }.toMono()
}
