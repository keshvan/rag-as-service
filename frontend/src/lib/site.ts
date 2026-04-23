export const appConfig = {
  name: "RAG Консоль",
  description: "Интерфейс для работы с корпоративными документами и умным поиском по знаниям.",
};

const pageMetaMap: Record<string, { title: string; subtitle: string }> = {
  "/": {
    title: "Дашборд",
    subtitle: "Обзор индексации, качества поиска и состояния системы.",
  },
  "/documents": {
    title: "Документы",
    subtitle: "Загрузка файлов и отслеживание статусов индексации.",
  },
  "/chat": {
    title: "RAG-чат",
    subtitle: "Вопросы к данным с указанием источников из индекса.",
  },
  "/settings": {
    title: "Настройки",
    subtitle: "Управление параметрами и доступом организации.",
  },
  "/profile": {
    title: "Профиль",
    subtitle: "Данные текущей сессии и быстрые действия для аккаунта.",
  },
  "/verification": {
    title: "Подтверждение email",
    subtitle: "Введите код из письма, чтобы завершить регистрацию и войти в систему.",
  },
};

export function resolvePageMeta(pathname: string) {
  return pageMetaMap[pathname] ?? pageMetaMap["/"];
}
