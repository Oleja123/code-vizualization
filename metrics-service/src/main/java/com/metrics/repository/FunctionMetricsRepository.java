package com.metrics.repository;

import com.metrics.entity.FunctionMetricsEntity;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface FunctionMetricsRepository extends JpaRepository<FunctionMetricsEntity, Long> {

    List<FunctionMetricsEntity> findByUsernameOrderByCreatedAtDesc(String username);

    long countByUsername(String username);

    List<FunctionMetricsEntity> findByUsernameAndCreatedAtBetween(
            String username, LocalDateTime from, LocalDateTime to);

    @Query("""
        SELECT f FROM FunctionMetricsEntity f
        WHERE f.username = :username
          AND f.createdAt = (
              SELECT MAX(f2.createdAt)
              FROM FunctionMetricsEntity f2
              WHERE f2.username = :username
          )
        """)
    List<FunctionMetricsEntity> findLatestByUsername(@Param("username") String username);
}