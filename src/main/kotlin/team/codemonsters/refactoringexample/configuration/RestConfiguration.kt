package team.codemonsters.refactoringexample.configuration

import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.context.annotation.Configuration

@Configuration
@ConfigurationProperties(prefix="rest")
data class RestConfiguration (
        var configs: Map<String, RestServiceCfg> = mutableMapOf()
)
