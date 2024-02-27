package team.codemonsters.refactoringexample.walletBalance

import org.slf4j.LoggerFactory
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpMethod
import org.springframework.http.HttpStatus
import org.springframework.http.MediaType
import org.springframework.stereotype.Component
import org.springframework.web.reactive.function.client.WebClient
import reactor.core.publisher.Mono
import reactor.kotlin.core.publisher.toMono
import java.net.URI

@Component
class RestGateway(private val webClient: WebClient
) {
    private val log = LoggerFactory.getLogger(Workflow::class.java)!!
    companion object {
        val CONTENT_TYPE = "${MediaType.APPLICATION_JSON};charset=${Charsets.UTF_8}"
    }

    fun sendRequest(request: RestServiceRequest): Mono<RestServiceResponse> {
        log.info("sending RestServiceRequest ${request}")
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
                            RestServiceResponse.of(
                                request,
                                HttpStatus.valueOf(response.statusCode().value()),
                                response.headers().asHttpHeaders(),
                                it
                            )
                        }
                }
                .onErrorResume { error ->
                    Mono.just(
                        RestServiceResponse.of(
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
                            RestServiceResponse.of(
                                request,
                                HttpStatus.valueOf(response.statusCode().value()),
                                response.headers().asHttpHeaders(),
                                it
                            )
                        }
                }
                .onErrorResume { error ->
                    Mono.just(
                        RestServiceResponse.of(
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
                    RestServiceResponse.of(
                        request,
                        HttpStatus.INTERNAL_SERVER_ERROR,
                        HttpHeaders.EMPTY,
                        "Ошибка при выполнении запроса к сервису ${request.restCfg.url}: метод ${request.requestEnvelope.method} не поддерживается"
                    )
                )
        }

    }

}