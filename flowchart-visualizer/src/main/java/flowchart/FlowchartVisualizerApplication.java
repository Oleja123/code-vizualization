package flowchart;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Главный класс Spring Boot приложения
 * Запускает REST API сервер на порту 8081
 */
@SpringBootApplication
public class FlowchartVisualizerApplication {
    
    public static void main(String[] args) {
        SpringApplication.run(FlowchartVisualizerApplication.class, args);
        
        System.out.println("\n===========================================");
        System.out.println("  Flowchart Visualizer API Started!");
        System.out.println("===========================================");
        System.out.println("  Server: http://localhost:8081");
        System.out.println("  Health: http://localhost:8081/api/flowchart/health");
        System.out.println("  API:    POST http://localhost:8081/api/flowchart/generate");
        System.out.println("===========================================\n");
    }
}
