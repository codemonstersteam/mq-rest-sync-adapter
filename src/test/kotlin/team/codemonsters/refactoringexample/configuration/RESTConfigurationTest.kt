package team.codemonsters.refactoringexample.configuration

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import team.codemonsters.refactoringexample.configuration.RestConfiguration


@SpringBootTest
class RESTConfigurationTest () {

    @Autowired
    val configurations: RestConfiguration?=null


    @Test
    fun restPropertiesTest(){

        Assertions.assertNotNull(configurations)
        Assertions.assertTrue(configurations!!.configs.isNotEmpty())

    }

    @Test
    fun restOperationsTest(){

        Assertions.assertNotNull(configurations)
        Assertions.assertTrue(configurations!!.configs.isNotEmpty())
        val pcrCfg = configurations!!.configs["api-gateway"]

        Assertions.assertTrue(pcrCfg!!.operations.isNotEmpty())
        Assertions.assertNotNull(pcrCfg.operations["client-reg"])
    }

}