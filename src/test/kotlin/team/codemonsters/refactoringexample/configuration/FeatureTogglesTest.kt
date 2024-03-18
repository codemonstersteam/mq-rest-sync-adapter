package team.codemonsters.refactoringexample.configuration

import org.junit.jupiter.api.Test
import org.assertj.core.api.Assertions.assertThat
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.test.context.ActiveProfiles

@SpringBootTest
@ActiveProfiles("test")
class FeatureTogglesTest (
    @Autowired val featureToggles: FeatureToggles
) {
    @Test
    fun `the non-existent feature must be disabled`(){
        assertThat(featureToggles.isFeatureEnabled("someFeature")).isFalse()
    }

    @Test
    fun `the test feature must be enabled`(){
        assertThat(featureToggles.isFeatureEnabled("test-feature")).isTrue()
    }

    @Test
    fun `the test feature must be disables`(){
        assertThat(featureToggles.isFeatureEnabled("test-feature-disabled")).isFalse()
    }

}