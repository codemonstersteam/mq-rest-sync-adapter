package team.codemonsters.refactoringexample.configuration

import org.springframework.stereotype.Component
import io.getunleash.DefaultUnleash

@Component
class FeatureToggles (private val featureToggles: DefaultUnleash) {
    
    fun isFeatureEnabled(name: String): Boolean =
        featureToggles.isEnabled(name)
    
}