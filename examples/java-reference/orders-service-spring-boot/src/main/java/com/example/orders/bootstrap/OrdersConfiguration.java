package com.example.orders.bootstrap;

import com.example.orders.application.CreateOrderUseCase;
import com.example.orders.infrastructure.InMemoryOrderRepository;
import com.example.orders.port.OrderRepository;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class OrdersConfiguration {
  @Bean
  public OrderRepository orderRepository() {
    return new InMemoryOrderRepository();
  }

  @Bean
  public CreateOrderUseCase createOrderUseCase(OrderRepository orderRepository) {
    return new CreateOrderUseCase(orderRepository);
  }
}
