package team.codemonsters.mqrestsyncadapter.test.stepdefs;

import java.io.File;
import java.nio.file.Files;

import javax.jms.Connection;
import javax.jms.Message;
import javax.jms.MessageConsumer;
import javax.jms.MessageProducer;
import javax.jms.Queue;
import javax.jms.Session;
import javax.jms.TextMessage;

import org.json.JSONObject;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotNull;

import com.ibm.mq.jms.MQQueueConnectionFactory;

import io.cucumber.java.en.And;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;
import team.codemonsters.mqrestsyncadapter.test.utils.TestConfig;

public class IbmMqSteps {
    private String payload;
    private Message sentMessage;
    private Message receivedMessage;
    private MQQueueConnectionFactory factory;
    private Connection connection;
    

    private void setupMqConnection() throws Exception {
        factory = new MQQueueConnectionFactory();
        factory.setHostName(TestConfig.MQ_HOST);  // Используем имя сервиса из docker-compose
        factory.setPort(TestConfig.MQ_PORT);
        factory.setQueueManager(TestConfig.MQ_QMGR_NAME);
        factory.setChannel(TestConfig.MQ_CHANNEL);  // Должен совпадать с MQ_CHANNEL в компоуз
        factory.setTransportType(1);  // Client transport type

        // Используем app пользователя, а не admin
        connection = factory.createConnection(TestConfig.MQ_USER, TestConfig.MQ_PASSWORD);
       
    }

    @Given("I have a valid request payload from {string}")
    public void loadPayload(String filename) throws Exception {
        File file = new File("src/test/resources/files/" + filename);
        payload = new String(Files.readAllBytes(file.toPath()));
    }

    @Given("IBM MQ server is running and accessible")
    public void verifyMqServerIsRunning() throws Exception {
        setupMqConnection();
    }

    @When("I send the message to queue {string}")
    public void sendMessage(String queueName) throws Exception {
        if (connection == null) {
            throw new IllegalStateException("MQ connection is not established. Make sure 'IBM MQ server is running and accessible' step is executed first.");
        }
        try (Session session = connection.createSession()) {
            Queue queue = session.createQueue(queueName);
            MessageProducer producer = session.createProducer(queue);

            TextMessage message = session.createTextMessage(payload);
            producer.send(message);
            sentMessage = message;
        }
    }

    @Then("the message should be successfully sent")
    public void verifyMessageSent() {
        assertNotNull("Message should be sent", sentMessage);
    }

    @When("I read the message from queue {string}")
    public void readMessage(String queueName) throws Exception {
        if (connection == null) {
            throw new IllegalStateException("MQ connection is not established. Make sure 'IBM MQ server is running and accessible' step is executed first.");
        }
        try (Session session = connection.createSession()) {
            Queue queue = session.createQueue(queueName);
            MessageConsumer consumer = session.createConsumer(queue);
            connection.start();
            receivedMessage = consumer.receive(5000);
        }
    }

    @Then("the received message should match the sent payload")
    public void verifyMessageContent() throws Exception {
        assertNotNull("Message should be received", receivedMessage);
        assertEquals("Message content should match", payload, ((TextMessage) receivedMessage).getText());
    }

    @And("the message should contain the correct correlation ID")
    public void verifyCorrelationId() throws Exception {
        JSONObject json = new JSONObject(payload);
        assertEquals("correlationId", json.getString("correlationId"));
    }

    @And("the message should contain the correct operation type")
    public void verifyOperationType() throws Exception {
        JSONObject json = new JSONObject(payload);
        assertEquals("client-wallet-balance", json.getString("operation"));
    }
}