package team.codemonsters.refactoringexample.configuration

import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.context.annotation.Configuration

@Configuration
@ConfigurationProperties(prefix = "webclient")
data class WebClientProperties(
    var connectTimeoutMills: Int = 10000,
    var readTimeoutMills: Long = 20000,
    var writeTimeoutMills: Long = 20000,
    var keepAlive: KeepAlive = KeepAlive(),
    var logging: Logging = Logging()
) {

    data class KeepAlive(
        var enabled: Boolean = false,
    )

    data class Logging(
        var enabled: Boolean = false,
        var defaultLevel: String = "DEBUG"
    )
}