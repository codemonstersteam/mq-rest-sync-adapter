package team.codemonsters.refactoringexample.configuration

import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.context.annotation.Configuration

@Configuration
@ConfigurationProperties(prefix="features")
data class FeatureTogglesConfiguration (
    var toggles: Map<String, FeatureToggle>
)

data class FeatureToggle(var enabled: Boolean, var description: String)
