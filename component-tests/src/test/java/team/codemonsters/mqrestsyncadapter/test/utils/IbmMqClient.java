package team.codemonsters.mqrestsyncadapter.test.utils;

import java.util.Map;
import java.util.concurrent.TimeUnit;

import javax.jms.Connection;
import javax.jms.MessageConsumer;
import javax.jms.MessageProducer;
import javax.jms.Queue;
import javax.jms.Session;
import javax.jms.TextMessage;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.ibm.mq.jms.MQQueueConnectionFactory;

public class IbmMqClient {
    private final ObjectMapper objectMapper;
    private final MQQueueConnectionFactory factory;
    private Connection connection;
    public static final int MAX_RETRIES = 5;
    public static final int RETRY_DELAY_MS = 100;

    public IbmMqClient() {
        this.objectMapper = new ObjectMapper();
        this.factory = new MQQueueConnectionFactory();
    }

    public void setupConnection() throws Exception {
        factory.setHostName(TestConfig.MQ_HOST);
        factory.setPort(TestConfig.MQ_PORT);
        factory.setQueueManager(TestConfig.MQ_QMGR_NAME);
        factory.setChannel(TestConfig.MQ_CHANNEL);
        factory.setTransportType(1);
        connection = factory.createConnection(TestConfig.MQ_USER, TestConfig.MQ_PASSWORD);
        connection.start();
    }

    public void sendMessage(String queueName, String payload, Map<String, String> properties) throws Exception {
        if (connection == null) {
            throw new java.lang.IllegalStateException("MQ connection is not established");
        }
        Session session = connection.createSession();
        Queue queue = session.createQueue(queueName);
        MessageProducer producer = session.createProducer(queue);

        TextMessage message = session.createTextMessage(payload);
        for (Map.Entry<String, String> entry : properties.entrySet()) {
            message.setStringProperty(entry.getKey(), entry.getValue());
        }
        
        producer.send(message);
        producer.close();
        session.close();
        TimeUnit.SECONDS.sleep(2);
    }

    public Map<String, Object> readMessage(String queueName, long timeout) throws Exception {
        if (connection == null) {
            throw new java.lang.IllegalStateException("MQ connection is not established");
        }

        TextMessage receivedMessage = null;
        int retryCount = 0;
        Session session = connection.createSession();
        Queue queue = session.createQueue(queueName);
        MessageConsumer consumer = session.createConsumer(queue);

        try {
            while (retryCount < MAX_RETRIES) {
                receivedMessage = (TextMessage) consumer.receive(timeout);
                if (receivedMessage != null) {
                    String messageText = receivedMessage.getText();
                    System.out.println("Attempt " + (retryCount + 1) + " - Received message: " + messageText);
                    if (messageText != null) {
                        return objectMapper.readValue(messageText, new TypeReference<>() {});
                    }
                }
                retryCount++;
                if (retryCount < MAX_RETRIES) {
                    TimeUnit.MILLISECONDS.sleep(RETRY_DELAY_MS);
                }
            }
        } finally {
            consumer.close();
            session.close();
        }

        if (receivedMessage == null) {
            throw new java.lang.IllegalStateException("Failed to receive message after " + MAX_RETRIES + " retries");
        }
        return null;
    }

    public void close() throws Exception {
        if (connection != null) {
            connection.close();
        }
    }
} 