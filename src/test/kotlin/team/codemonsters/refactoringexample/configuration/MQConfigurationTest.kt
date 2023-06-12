package team.codemonsters.refactoringexample.configuration

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import team.codemonsters.refactoringexample.configuration.MqConfiguration
import team.codemonsters.refactoringexample.configuration.MqServiceCfg


@SpringBootTest
class MQConfigurationTest() {

    @Autowired
    val configurations: MqConfiguration? = null

    @Autowired
    val esiaConf: MqServiceCfg? = null

    @Test
    fun mqPropertiesTest() {

        Assertions.assertNotNull(configurations)
        Assertions.assertTrue(configurations!!.configs.isNotEmpty())

    }
}