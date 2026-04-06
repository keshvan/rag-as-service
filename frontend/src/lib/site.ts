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
};

export function resolvePageMeta(pathname: string) {
  return pageMetaMap[pathname] ?? pageMetaMap["/"];
}
