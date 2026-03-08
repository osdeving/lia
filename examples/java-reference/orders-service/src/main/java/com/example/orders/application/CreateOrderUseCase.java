package com.example.orders.application;

import com.example.orders.domain.OrderStatus;
import com.example.orders.port.OrderRepository;

public final class CreateOrderUseCase {
  private final OrderRepository orderRepository;

  public CreateOrderUseCase(OrderRepository orderRepository) {
    this.orderRepository = orderRepository;
  }

  public CreateOrderOutput execute(CreateOrderInput input) {
    var stored = orderRepository.save(input.id());
    return new CreateOrderOutput(stored, OrderStatus.NEW);
  }
}
