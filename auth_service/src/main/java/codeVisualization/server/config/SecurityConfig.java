package codeVisualization.server.config;

import codeVisualization.server.service.CustomUserDetailsService;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.config.annotation.authentication.builders.AuthenticationManagerBuilder;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;

import java.util.List;

@Configuration
public class SecurityConfig {

    private final CustomUserDetailsService userDetailsService;

    public SecurityConfig(CustomUserDetailsService userDetailsService) {
        this.userDetailsService = userDetailsService;
    }

    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {

        http.cors(cors -> cors.configurationSource(corsConfigurationSource()))
                .csrf(csrf -> csrf.disable())

                // ВАЖНО: Настройка сессий
                .sessionManagement(session -> session
                        .sessionCreationPolicy(SessionCreationPolicy.IF_REQUIRED)
                        .maximumSessions(1)
                )

                .authorizeHttpRequests(auth -> auth
                        // Статика
                        .requestMatchers("/css/**", "/js/**", "/images/**").permitAll()

                        // Страницы и API авторизации - доступны всем
                        .requestMatchers("/login.html", "/register.html").permitAll()
                        .requestMatchers("/api/auth/**").permitAll()  // Все endpoints /api/auth/*

                        // Всё остальное API — только авторизованным
                        .requestMatchers("/api/**").authenticated()

                        // Все остальные запросы
                        .anyRequest().permitAll()
                )

                // УБИРАЕМ formLogin - он мешает JSON авторизации
                // .formLogin() - закомментировано!

                // УБИРАЕМ стандартный logout
                // .logout() - закомментировано!

                .exceptionHandling(exception -> exception
                        .authenticationEntryPoint((request, response, authException) -> {
                            response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
                            response.getWriter().write("Unauthorized");
                        })
                );

        return http.build();
    }

    @Bean
    public AuthenticationManager authManager(HttpSecurity http) throws Exception {
        var builder = http.getSharedObject(AuthenticationManagerBuilder.class);
        builder.userDetailsService(userDetailsService)
                .passwordEncoder(passwordEncoder());
        return builder.build();
    }

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }

    @Bean
    public CorsConfigurationSource corsConfigurationSource() {
        CorsConfiguration config = new CorsConfiguration();

        // Разрешаем все возможные порты фронтенда
        config.setAllowedOriginPatterns(List.of(
                "http://localhost:3000",     // React/Next.js
                "http://localhost:5173",     // Vite
                "http://localhost:8080",     // Vue CLI
                "http://localhost:4200",     // Angular
                "http://localhost:8082"      // Другие сервисы
        ));

        config.setAllowedMethods(List.of("GET", "POST", "PUT", "DELETE", "OPTIONS"));
        config.setAllowedHeaders(List.of("*"));
        config.setAllowCredentials(true);  // ОБЯЗАТЕЛЬНО для cookies/сессий!
        config.setExposedHeaders(List.of("Set-Cookie"));  // Разрешаем браузеру видеть Set-Cookie

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", config);
        return source;
    }
}