package team.codemonsters.refactoringexample.service;

import org.slf4j.LoggerFactory
import org.springframework.messaging.Message
import org.springframework.stereotype.Component
import reactor.core.publisher.Mono
import team.codemonsters.refactoringexample.client.RESTClient

import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.mq.HttpRequestEnvelope.Companion.fromJsonString
import team.codemonsters.refactoringexample.mq.enum.ReceiptStatus
import team.codemonsters.refactoringexample.mq.*
import team.codemonsters.refactoringexample.util.JsonUtil
import java.util.*

/**
 * Пайплайн запроса АС для получения историй операций по кошельку
 *
 * - Получает конверт с запросом от АС
 * - Валидирует входящий запрос
 * - Формирует квитанцию
 * - Отправляет квитанцию ACCEPTED в случае успешной валидации
 * - Формирует запрос для получения истории операций
 * - Отправляет запрос /wallet/balance
 * - Валидирует ответ
 * - Формирует квитацию
 * - Отправляет квитанцию SUCCESS в случае успешного ответа
 * - Отправляет ответ с выпиской по истории операций
 */

@Component
class PipelineServicePS() : PipelineService {

    private val log = LoggerFactory.getLogger(MqHttpService::class.java)!!

    private lateinit var receivedMessage: Message<Any>
    private lateinit var mqPublisher: MqPublisher
    private lateinit var mqConfig: MqServiceCfg
    private lateinit var restConfiguration: RestConfiguration
    private lateinit var restClient: RESTClient

    override fun receiveMessage(
        message: Message<Any>,
        mqPublisher: MqPublisher,
        mqConfig: MqServiceCfg,
        restConfiguration: RestConfiguration,
        restClient: RESTClient
    ) {
        this.receivedMessage = message
        this.mqPublisher = mqPublisher
        this.mqConfig = mqConfig
        this.restConfiguration = restConfiguration
        this.restClient = restClient

        log.info("Incoming transmission into MQ-REST-ADAPTER ")
        log.info("headers: ${message.headers}")
        log.info("payload: ${message.payload}")

        processMessage(message)
    }

    fun processMessage(message: Message<Any>) {
        val correlationId = message.headers[mqConfig.correlationID] as String? ?: UUID.randomUUID().toString()
        val url = message.headers["X_Method"] as String? ?: "/somemethod"

        log.debug("msg class is: ${message.javaClass}")

        // TODO - Нужно найти способ получать msgId через стандартные методы
        val msgId = message.headers["jms_messageId"] as String?
        var httpReq: String = message.payload.toString()


        Mono.just(
            HttpRequestEnvelope.fromJsonString(message.payload.toString())
        ).map {
            HttpRequestEnvelope.validatedRequest(it, mqConfig.timeout)
        }.map {
            HttpRequestEnvelope.putPsRequestInBody(it)
        }.map {
            acceptedMqReceipt(correlationId, url, it, msgId, mqPublisher, mqConfig, httpReq)
        }.map {
            RestServiceRequest.emerge(it, restConfiguration)
        }.flatMap {
            submitRestRequest(it, restClient)
        }.map {
            processMqResponse(it)
        }.map {
            MqResponseMessage.emerge(it)
        }.map {
            submitMqReceipt(correlationId, url, it, msgId, mqPublisher, mqConfig, httpReq)
        }.flatMap {
            submitMqResponse(correlationId, url, it, msgId, mqPublisher, mqConfig)
        }.block()
    }

    fun acceptedMqReceipt(
        correlationId: String,
        url: String,
        data: Result<HttpRequestEnvelope>,
        originalMsgId: String?,
        mqPublisher: MqPublisher,
        mqConfig: MqServiceCfg,
        httpRequestEnvelope: String
    ): Result<HttpRequestEnvelope> =
        data.fold(
            onSuccess = { success ->
                createMessageForReceipt(httpRequestEnvelope, ReceiptStatus.ACCEPTED, null, null)
                    .map { it ->
                        val receiptResponse = MqResponseMessage.emerge(correlationId, url, it)
                        val jmsResponse = convertToReceiptJmsResponse(receiptResponse, mqConfig)
                        mqPublisher.sendToMq(jmsResponse, originalMsgId)
                        return Result.success(success)
                    }

            },
            onFailure = { error ->
                Result.failure(error)
            }
        )

    fun createMessageForReceipt(
        httpRequestEnvelope: String,
        flag: ReceiptStatus,
        errorCode: Int?,
        description: String?
    ): Result<String> =
        fromJsonString(httpRequestEnvelope).fold(
            onSuccess = { request ->
                PsGetBalanceRequest.fromJsonString(request.body!!).fold(
                    onSuccess = { request ->
                        Result.success(
                            JsonUtil.toJson(
                                ReceiptPs(
                                    request.requestType,
                                    request.dateFrom,
                                    request.dateTo,
                                    request.operTime,
                                    flag,
                                    errorCode,
                                    description
                                )
                            )
                        )
                    },
                    onFailure = { error -> Result.failure(error) }
                )

            },
            onFailure = { error -> Result.failure(error) }
        ).map { it }


    private fun convertToReceiptJmsResponse(
        mqRequest: MqResponseMessage,
        mqConfig: MqServiceCfg
    ) = BasicJmsRequest(
        correlationId = mqRequest.correlationId,
        queue = mqConfig.RECIEPT,
        payload = mqRequest.payload,
        headers = prepareSipHeaders(mqRequest.correlationId, mqRequest.operation, mqConfig)
    )

    private fun convertToBasicJmsResponse(
        mqRequest: MqResponseMessage,
        mqConfig: MqServiceCfg
    ) = BasicJmsRequest(
        correlationId = mqRequest.correlationId,
        queue = mqConfig.RES,
        payload = mqRequest.payload,
        headers = prepareSipHeaders(mqRequest.correlationId, mqRequest.operation, mqConfig)
    )

    private fun prepareSipHeaders(correlationId: String, url: String, mqConfig: MqServiceCfg): Map<String, String> {
        val headers = mutableMapOf(
            mqConfig.correlationID to correlationId,
            "X_Method" to url,
            "RequestID" to UUID.randomUUID().toString(),
            mqConfig.createDateTime to mqConfig.getFormattedCreateDateTime(),
        )
        headers.putAll(mqConfig.values)
        return headers
    }

    fun submitRestRequest(data: Result<RestServiceRequest>, restClient: RESTClient): Mono<Result<RestServiceResponse>> =
        data.fold(
            onSuccess = { success -> restClient.sendRequest(success) },
            onFailure = { error -> Mono.just(Result.failure(error)) }
        )

    fun processMqResponse(
        data: Result<RestServiceResponse>,
    ): Result<HttpResponseEnvelope> =
        data.fold(
            onSuccess = { success -> HttpResponseEnvelope.emerge(success) },
            onFailure = { error ->
                log.debug(error.message)
                Result.failure(error)
            })

    fun submitMqReceipt(
        correlationId: String,
        url: String,
        mqResponse: Result<MqResponseMessage>,
        originalMessageId: String?,
        mqPublisher: MqPublisher,
        mqConfig: MqServiceCfg,
        httpRequestEnvelope: String
    ): Result<MqResponseMessage> =
        mqResponse.fold(
            onSuccess = { success ->
                createMessageForReceipt(httpRequestEnvelope, ReceiptStatus.SUCCESS, null, null).map {
                    val receiptResponse = MqResponseMessage.emerge(correlationId, url, it)
                    val jmsResponse = convertToReceiptJmsResponse(receiptResponse, mqConfig)
                    mqPublisher.sendToMq(jmsResponse, originalMessageId)
                    return Result.success(success)
                }
            },
            onFailure = { error ->
                createMessageForReceipt(httpRequestEnvelope, ReceiptStatus.ERROR, null, error.message).map {
                    val receiptResponse = MqResponseMessage.emerge(correlationId, url, it)
                    val jmsResponse = convertToReceiptJmsResponse(receiptResponse, mqConfig)
                    mqPublisher.sendToMq(jmsResponse, originalMessageId)
                    return Result.failure(error)
                }
            }
        )

    fun submitMqResponse(
        correlationId: String,
        url: String,
        data: Result<MqResponseMessage>,
        originalMsgId: String?,
        mqPublisher: MqPublisher,
        mqConfig: MqServiceCfg
    ): Mono<MqPublishedResponce<String>> =
        data.fold(
            onSuccess = { success ->
                mqPublisher.publishToMq(
                    convertToBasicJmsResponse(success, mqConfig),
                    originalMsgId
                )
            },
            onFailure = { error ->
                mqPublisher.publishToMq(
                    convertToBasicJmsResponse(
                        MqResponseMessage.emerge(
                            correlationId,
                            url,
                            error
                        ), mqConfig
                    ), originalMsgId
                )
            }
        )
}
