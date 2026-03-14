package codeVisualization.server.controller;

import codeVisualization.server.db.jpaRepository.UserRepository;
import codeVisualization.server.model.UsersDto;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpSession;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContext;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.context.HttpSessionSecurityContextRepository;
import org.springframework.web.bind.annotation.*;
import codeVisualization.server.db.entity.User;

import java.util.Collections;
import java.util.Optional;

@Slf4j
@RestController
@RequestMapping("/api/auth")
@RequiredArgsConstructor
public class AuthController {

    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;

    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody UsersDto userDto, HttpServletRequest request) {
        Optional<User> user = userRepository.findByName(userDto.getUsername());

        log.info("Попытка входа: пользователь '{}'", userDto.getUsername());

        if (user.isEmpty() || !passwordEncoder.matches(userDto.getRawPassword(), user.get().getPassword())) {
            log.warn("Ошибка входа: пользователь '{}' ввёл неверные данные", userDto.getUsername());
            return ResponseEntity.status(401).body("Неверный логин или пароль");
        }

        // ВАЖНО: Создаем аутентификацию и сохраняем в SecurityContext
        UsernamePasswordAuthenticationToken authToken =
                new UsernamePasswordAuthenticationToken(userDto.getUsername(), null, Collections.emptyList());

        SecurityContext securityContext = SecurityContextHolder.getContext();
        securityContext.setAuthentication(authToken);

        // ВАЖНО: Сохраняем SecurityContext в сессию
        HttpSession session = request.getSession(true);
        session.setAttribute(HttpSessionSecurityContextRepository.SPRING_SECURITY_CONTEXT_KEY, securityContext);

        log.info("Успешный вход пользователя '{}', сессия ID: {}", userDto.getUsername(), session.getId());

        return ResponseEntity.ok("Успешный вход");
    }

    @PostMapping("/logout")
    public ResponseEntity<?> logout(HttpServletRequest request) {
        HttpSession session = request.getSession(false);
        if (session != null) {
            log.info("Выход пользователя, сессия ID: {}", session.getId());
            session.invalidate();
        }
        SecurityContextHolder.clearContext();
        return ResponseEntity.ok("Вы вышли из системы");
    }

    @GetMapping("/me")
    public ResponseEntity<?> currentUser(HttpServletRequest request) {
        Authentication authentication = SecurityContextHolder.getContext().getAuthentication();

        log.debug("Проверка сессии. Auth: {}, Principal: {}",
                authentication != null ? authentication.getClass().getSimpleName() : "null",
                authentication != null ? authentication.getPrincipal() : "null");

        if (authentication == null
                || !authentication.isAuthenticated()
                || authentication.getPrincipal().equals("anonymousUser")) {

            log.debug("Пользователь не авторизован");
            return ResponseEntity.status(401).body("Не авторизован");
        }

        String username = authentication.getName();
        log.info("Сессия активна для пользователя '{}'", username);
        return ResponseEntity.ok(username);
    }

    @PostMapping("/register")
    public ResponseEntity<?> register(@RequestBody UsersDto userDto, HttpServletRequest request) {
        log.info("Попытка регистрации: пользователь '{}'", userDto.getUsername());

        if (userRepository.findByName(userDto.getUsername()).isPresent()) {
            log.warn("Пользователь '{}' уже существует", userDto.getUsername());
            return ResponseEntity
                    .status(HttpStatus.CONFLICT)
                    .body("Пользователь уже существует");
        }

        User user = new User();
        user.setName(userDto.getUsername());
        user.setPassword(passwordEncoder.encode(userDto.getRawPassword()));
        userRepository.save(user);

        log.info("Пользователь '{}' успешно зарегистрирован", userDto.getUsername());

        // После регистрации НЕ логиним автоматически
        // Пользователь должен сам войти
        return ResponseEntity.status(HttpStatus.CREATED).body("Пользователь создан");
    }

    @GetMapping("/")
    public String home(Authentication auth) {
        if (auth == null || !auth.isAuthenticated() || "anonymousUser".equals(auth.getPrincipal())) {
            return "redirect:/login.html";
        }
        return "index.html";
    }
}