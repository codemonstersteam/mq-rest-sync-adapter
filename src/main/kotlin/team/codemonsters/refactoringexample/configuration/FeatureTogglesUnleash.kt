package team.codemonsters.refactoringexample.configuration

import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import io.getunleash.util.UnleashConfig
import io.getunleash.event.UnleashSubscriber
import io.getunleash.event.UnleashReady
import io.getunleash.repository.FeatureToggleResponse
import io.getunleash.repository.ToggleCollection
import io.getunleash.DefaultUnleash

@Configuration
@ConfigurationProperties(prefix = "unleash")
data class FeatureTogglesUnleash(
    var appName: String = "",
    var instanceId: String = "",
    var url: String = "",
    var token: String = "",
    var environment: String = "",
) {

    @Bean
    fun unleashConfiguration(): UnleashConfig =
        UnleashConfig
            .builder()
            .appName(appName)
            .instanceId(instanceId)
            .unleashAPI(url)
            .apiKey(token)
            .environment(environment)
            .build()

    
    @Bean
    fun unleash(): DefaultUnleash = DefaultUnleash(unleashConfiguration())
}
