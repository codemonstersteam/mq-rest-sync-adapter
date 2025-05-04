package team.codemonsters.refactoringexample.configuration

import org.springframework.stereotype.Component

@Component
class FeatureToggles (private val configuration: FeatureTogglesConfiguration) {
    
    fun isFeatureEnabled(name: String): Boolean =
        configuration.toggles.contains(name) && configuration.toggles[name]!!.enabled
    
}