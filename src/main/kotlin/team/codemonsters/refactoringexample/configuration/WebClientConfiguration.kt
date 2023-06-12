package team.codemonsters.refactoringexample.configuration

import ch.qos.logback.classic.Level
import ch.qos.logback.classic.Logger
import io.netty.channel.ChannelOption
import io.netty.handler.logging.LogLevel
import io.netty.handler.timeout.ReadTimeoutHandler
import io.netty.handler.timeout.WriteTimeoutHandler
import org.slf4j.LoggerFactory
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.http.client.reactive.ReactorClientHttpConnector
import org.springframework.web.reactive.function.client.WebClient
import reactor.netty.http.client.HttpClient
import reactor.netty.transport.logging.AdvancedByteBufFormat
import java.util.concurrent.TimeUnit

@Configuration
class WebClientConfiguration(private val webClientProperties: WebClientProperties) {

    @Bean
    fun webClient(): WebClient = WebClient.builder()
        .clientConnector(ReactorClientHttpConnector(buildHttpClient()))
        .build()

    private fun buildHttpClient(): HttpClient = enableLogging(HttpClient.create())
        .compress(true)
        .option(ChannelOption.CONNECT_TIMEOUT_MILLIS, webClientProperties.connectTimeoutMills)
        .option(ChannelOption.SO_KEEPALIVE, webClientProperties.keepAlive.enabled)
        .doOnConnected {
            it.addHandlerLast(ReadTimeoutHandler(webClientProperties.readTimeoutMills, TimeUnit.MILLISECONDS))
            it.addHandlerLast(WriteTimeoutHandler(webClientProperties.writeTimeoutMills, TimeUnit.MILLISECONDS))
        }

    private fun enableLogging(httpClient: HttpClient): HttpClient = httpClient.let {
        if (webClientProperties.logging.enabled)
            it.wiretap(HttpClient::class.java.name, LogLevel.valueOf(getLogLevel()), AdvancedByteBufFormat.TEXTUAL)
        else
            it.wiretap(false)
    }

    private fun getLogLevel(): String {
        val logger = LoggerFactory.getLogger(HttpClient::class.java.packageName) as Logger
        return logger.run {
            if (level == null) level = Level.toLevel(webClientProperties.logging.defaultLevel.toUpperCase())
            level.levelStr
        }
    }
}