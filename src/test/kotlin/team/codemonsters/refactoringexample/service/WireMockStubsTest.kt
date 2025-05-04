package team.codemonsters.refactoringexample.service

import com.github.tomakehurst.wiremock.WireMockServer
import com.maciejwalkowiak.wiremock.spring.ConfigureWireMock
import com.maciejwalkowiak.wiremock.spring.EnableWireMock
import com.maciejwalkowiak.wiremock.spring.InjectWireMock
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.test.context.ActiveProfiles
import org.springframework.test.web.reactive.server.WebTestClient

@ActiveProfiles("test")
@SpringBootTest
@EnableWireMock(ConfigureWireMock(name = "rest-api-gateway", property = "rest.configs.api-gateway.url"))
class WireMockStubsTest {
    @InjectWireMock("rest-api-gateway")
    private lateinit var restApiGateway: WireMockServer
    lateinit var client: WebTestClient

    @BeforeEach
    fun setUp() {
        client = WebTestClient.bindToServer().baseUrl(restApiGateway.baseUrl()).build()
    }

    @Test
    fun wireMockSuccess() {
        client.post().uri("/api/client/Api-clientId-01/wallet/balance")
            .bodyValue("{}")
            .exchange()
            .expectStatus().isOk
    }

}