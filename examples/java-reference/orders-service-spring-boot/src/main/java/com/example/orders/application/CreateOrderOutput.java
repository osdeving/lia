package com.example.orders.application;

import com.example.orders.domain.OrderStatus;

public record CreateOrderOutput(boolean accepted, OrderStatus status) {
}
