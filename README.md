# testauth

Для первого запуска testauth -init это создаст необходимые таблицы и роль администратора

**таблицы:**
* dictionaries
* accounts
* roles
* checkpoints
* predicates
* связующие таблицы

Аккаунты, роли, контрольные точки, предикаты и справочники - все является справочниками и для каждого имеется запись в таблице **dictionary** 

Для каждого справочника создается одна контрольная точка на все CRUD-операции

Все создаваемые контрольные точки добавляются в роль админа

Связь с контрольными точками имеется как у ролей так и у пользователей, то же самое и с предикатами

`**POST /signup** email, password, confirm` при первом запуске создаст пользователя с ролью администратора, нужно создать второго пользователя, чтобы посмотреть на поведение хендлеров для него

`**GET /login** email, password` вернет токен
