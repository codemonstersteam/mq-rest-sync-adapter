package team.codemonsters.refactoringexample.client

import org.springframework.http.HttpHeaders
import org.springframework.http.HttpMethod
import org.springframework.http.HttpStatus
import org.springframework.http.MediaType
import org.springframework.stereotype.Component
import org.springframework.web.reactive.function.client.WebClient
import reactor.core.publisher.Mono
import reactor.kotlin.core.publisher.toMono

import team.codemonsters.refactoringexample.mq.RestServiceRequest
import team.codemonsters.refactoringexample.mq.RestServiceResponse
import java.net.URI

/**
 * Простой REST клиент - просто пробрасывает запрос в сервис и возвращает ответ в виде RestServiceResponse
 * В случае ошибки при передаче данных возвращает RestServiceResponse с кодом 500
 */
@Component
class RESTClient(private val webClient: WebClient
) {

    private val SERVICE = "mq-rest-sync-adapter"

    fun sendRequest(request: RestServiceRequest): Mono<Result<RestServiceResponse>> {

        when (request.requestEnvelope.method) {
            //Post request
            HttpMethod.POST.name() -> return webClient.post()
                .uri(URI.create(request.restCfg.url + request.operationUrl))
                .headers {
                    it.set(HttpHeaders.CONTENT_TYPE, CONTENT_TYPE)
                    request.requestEnvelope.headers.forEach { (k, v) ->
                        it.set(k, v)
                    }
                    if (request.restCfg.basicAuth != null) {
                        it.setBasicAuth(request.restCfg.basicAuth!!.user, request.restCfg.basicAuth!!.password)
                    }
                }
                .bodyValue(request.requestEnvelope.body ?: "")
                .exchangeToMono { response ->
                    response.bodyToMono(String::class.java)
                        .switchIfEmpty("".toMono())
                        .map {
                            RestServiceResponse.emerge(
                                request,
                                HttpStatus.valueOf(response.statusCode().value()),
                                response.headers().asHttpHeaders(),
                                it
                            )
                        }
                }
                .onErrorResume { error ->
                    Mono.just(
                        RestServiceResponse.emerge(
                            request,
                            HttpStatus.INTERNAL_SERVER_ERROR,
                            HttpHeaders.EMPTY,
                            "Ошибка ${error.message} при выполнении запроса к сервису ${request.restCfg.url}"
                        )
                    )
                }
            // Get request
            HttpMethod.GET.name() -> return webClient.get()
                .uri(URI.create(request.restCfg.url + request.operationUrl))
                .headers {
                    it.set(HttpHeaders.CONTENT_TYPE, CONTENT_TYPE)
                    request.requestEnvelope.headers.forEach { (k, v) ->
                        it.set(k, v)
                    }
                    if (request.restCfg.basicAuth != null) {
                        it.setBasicAuth(request.restCfg.basicAuth!!.user, request.restCfg.basicAuth!!.password)
                    }
                }
                .exchangeToMono { response ->
                    response.bodyToMono(String::class.java)
                        .switchIfEmpty("".toMono())
                        .map {
                            RestServiceResponse.emerge(
                                request,
                                HttpStatus.valueOf(response.statusCode().value()),
                                response.headers().asHttpHeaders(),
                                it
                            )
                        }
                }
                .onErrorResume { error ->
                    Mono.just(
                        RestServiceResponse.emerge(
                            request,
                            HttpStatus.INTERNAL_SERVER_ERROR,
                            HttpHeaders.EMPTY,
                            "Ошибка ${error.message} при выполнении запроса к сервису ${request.restCfg.url}"
                        )
                    )
                }
            // Остальные типы запросов не поддерживаются
            else ->
                return Mono.just(
                    RestServiceResponse.emerge(
                        request,
                        HttpStatus.INTERNAL_SERVER_ERROR,
                        HttpHeaders.EMPTY,
                        "Ошибка при выполнении запроса к сервису ${request.restCfg.url}: метод ${request.requestEnvelope.method} не поддерживается"
                    )
                )
        }

    }
    companion object {
        val CONTENT_TYPE = "${MediaType.APPLICATION_JSON};charset=${Charsets.UTF_8}"
    }
}
