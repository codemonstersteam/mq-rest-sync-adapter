package team.codemonsters.refactoringexample.domain

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Test
import org.springframework.boot.test.context.SpringBootTest
import team.codemonsters.refactoringexample.TestUtil
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.configuration.RestServiceCfg
import team.codemonsters.refactoringexample.mq.HttpRequestEnvelope
import team.codemonsters.refactoringexample.mq.RestServiceRequest


@SpringBootTest
class StringSubstitutionTest {

    @Test
    fun replaceVariables(){
        val httpRequestEnvelope = Result.success(TestUtil.readResourceFilePathTo("/data/domain/client-cert-reg-request.json", HttpRequestEnvelope::class.java))

        val restConfiguration = RestConfiguration(
            configs = mapOf("cifrub-processing" to RestServiceCfg(
                url = "http://base-service-utl:8080",
                basicAuth = null,
                operations = mapOf("client-cert-reg" to "/api/v2.0/client/\${Api-clientId}/certificate/register")
            )
            )
        )

        val httpRequest = RestServiceRequest.emerge(httpRequestEnvelope,restConfiguration )

        Assertions.assertNotNull(httpRequest.getOrNull())
        Assertions.assertEquals("/api/v2.0/client/aaa-bbb/certificate/register", httpRequest.getOrNull()!!.operationUrl)
    }
}