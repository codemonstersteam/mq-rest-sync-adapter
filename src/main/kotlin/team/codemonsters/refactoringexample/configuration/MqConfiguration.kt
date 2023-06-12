package team.codemonsters.refactoringexample.configuration

import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.context.annotation.Configuration

@Configuration
@ConfigurationProperties(prefix="mq")
data class MqConfiguration (
        var configs: Map<String, MqServiceCfg> = mutableMapOf()
)
