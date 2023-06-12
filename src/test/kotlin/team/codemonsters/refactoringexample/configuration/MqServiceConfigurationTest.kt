package team.codemonsters.refactoringexample.configuration

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.context.ApplicationContext
import team.codemonsters.refactoringexample.configuration.MqConfiguration
import team.codemonsters.refactoringexample.service.MqHttpService
import team.codemonsters.refactoringexample.service.MqServiceConfiguration


@SpringBootTest
class MqServiceConfigurationTest (    @Autowired val applicationContext: ApplicationContext,
    @Autowired val configurations: MqConfiguration
) {


    @Test
    fun mqServicesTest(){

        val serviceBeans = applicationContext.getBeansOfType(MqHttpService::class.java)
        Assertions.assertTrue(serviceBeans.isNotEmpty())

        configurations.configs.forEach { (t, u) ->
            Assertions.assertTrue(serviceBeans.containsKey(MqServiceConfiguration.getBeanName(t)))
        }

    }
}