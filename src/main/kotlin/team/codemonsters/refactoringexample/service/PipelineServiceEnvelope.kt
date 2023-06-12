package team.codemonsters.refactoringexample.service

import org.slf4j.LoggerFactory
import org.springframework.messaging.Message
import org.springframework.stereotype.Component
import reactor.core.publisher.Mono
import team.codemonsters.refactoringexample.client.RESTClient
import team.codemonsters.refactoringexample.configuration.MqServiceCfg
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.mq.*
import java.util.*


/**
 * Пайплайн обработки входящих запросов от MIDDLEWARE
 *
 * - Получает конверт с запросом
 * - Формирует запрос
 * - Отправляет запрос в выбранный сервис
 * - Валидирует ответ
 * - Формирует ответное сообщение
 * - Отправляет сообщение
 */


@Component
class PipelineServiceEnvelope() : PipelineService {

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
        restClient: RESTClient,
    ) {
        this.receivedMessage = message
        this.mqPublisher = mqPublisher
        this.mqConfig = mqConfig
        this.restConfiguration = restConfiguration
        this.restClient = restClient

        log.info("Incoming transmission into DRR ")
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

        Mono.just(
            HttpRequestEnvelope.fromJsonString(message.payload.toString())
        ).map {
            RestServiceRequest.emerge(it, restConfiguration)
        }.flatMap {
            submitRestRequest(it, restClient)
        }.map {
            HttpResponseEnvelope.emerge(it)
        }.map {
            MqResponseMessage.emerge(it)
        }.flatMap {
            submitMqResponse(correlationId, url, it, msgId, mqPublisher, mqConfig)
        }.block()
    }

    fun submitRestRequest(data: Result<RestServiceRequest>, restClient: RESTClient): Mono<Result<RestServiceResponse>> =
        data.fold(
            onSuccess = { success -> restClient.sendRequest(success) },
            onFailure = { error ->
                log.debug(error.message)
                Mono.just(Result.failure(error))
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
            onSuccess = { success -> mqPublisher.publishToMq(convertToBasicJmsResponse(success, mqConfig), originalMsgId) },
            onFailure = { error -> mqPublisher.publishToMq(convertToBasicJmsResponse(MqResponseMessage.emerge(correlationId, url, error), mqConfig), originalMsgId) }
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
    
}