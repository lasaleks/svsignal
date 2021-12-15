-- phpMyAdmin SQL Dump
-- version 5.1.1
-- https://www.phpmyadmin.net/
--
-- Хост: mysql
-- Время создания: Дек 15 2021 г., 04:29
-- Версия сервера: 10.5.8-MariaDB-1:10.5.8+maria~focal-log
-- Версия PHP: 7.4.20

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- База данных: `insiteexpert_v4`
--

-- --------------------------------------------------------

--
-- Структура таблицы `svsignal_group`
--

CREATE TABLE `svsignal_group` (
  `id` int(11) NOT NULL,
  `group_key` varchar(200) NOT NULL,
  `name` varchar(200) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `svsignal_group`
--
ALTER TABLE `svsignal_group`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `svsignal_group`
--
ALTER TABLE `svsignal_group`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;



-- phpMyAdmin SQL Dump
-- version 5.1.1
-- https://www.phpmyadmin.net/
--
-- Хост: mysql
-- Время создания: Дек 15 2021 г., 04:30
-- Версия сервера: 10.5.8-MariaDB-1:10.5.8+maria~focal-log
-- Версия PHP: 7.4.20

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- База данных: `insiteexpert_v4`
--

-- --------------------------------------------------------

--
-- Структура таблицы `svsignal_signal`
--

CREATE TABLE `svsignal_signal` (
  `id` int(11) NOT NULL,
  `group_id` int(11) NOT NULL,
  `signal_key` varchar(200) NOT NULL,
  `name` varchar(200) NOT NULL,
  `type_save` int(11) NOT NULL,
  `period` int(11) NOT NULL,
  `delta` float NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `svsignal_signal`
--
ALTER TABLE `svsignal_signal`
  ADD PRIMARY KEY (`id`),
  ADD KEY `group_id` (`group_id`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `svsignal_signal`
--
ALTER TABLE `svsignal_signal`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `svsignal_signal`
--
ALTER TABLE `svsignal_signal`
  ADD CONSTRAINT `svsignal_signal_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `svsignal_group` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;


-- phpMyAdmin SQL Dump
-- version 5.1.1
-- https://www.phpmyadmin.net/
--
-- Хост: mysql
-- Время создания: Дек 15 2021 г., 04:30
-- Версия сервера: 10.5.8-MariaDB-1:10.5.8+maria~focal-log
-- Версия PHP: 7.4.20

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- База данных: `insiteexpert_v4`
--

-- --------------------------------------------------------

--
-- Структура таблицы `svsignal_tag`
--

CREATE TABLE `svsignal_tag` (
  `id` int(11) NOT NULL,
  `signal_id` int(11) NOT NULL,
  `tag` varchar(200) NOT NULL,
  `value` varchar(400) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `svsignal_tag`
--
ALTER TABLE `svsignal_tag`
  ADD PRIMARY KEY (`id`),
  ADD KEY `signal_id` (`signal_id`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `svsignal_tag`
--
ALTER TABLE `svsignal_tag`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `svsignal_tag`
--
ALTER TABLE `svsignal_tag`
  ADD CONSTRAINT `svsignal_tag_ibfk_1` FOREIGN KEY (`signal_id`) REFERENCES `svsignal_signal` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;


-- phpMyAdmin SQL Dump
-- version 5.1.1
-- https://www.phpmyadmin.net/
--
-- Хост: mysql
-- Время создания: Дек 15 2021 г., 04:31
-- Версия сервера: 10.5.8-MariaDB-1:10.5.8+maria~focal-log
-- Версия PHP: 7.4.20

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- База данных: `insiteexpert_v4`
--

-- --------------------------------------------------------

--
-- Структура таблицы `svsignal_ivalue`
--

CREATE TABLE `svsignal_ivalue` (
  `id` int(11) NOT NULL,
  `signal_id` int(11) NOT NULL,
  `utime` int(11) NOT NULL,
  `value` int(11) NOT NULL,
  `offline` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `svsignal_ivalue`
--
ALTER TABLE `svsignal_ivalue`
  ADD PRIMARY KEY (`id`),
  ADD KEY `signal_id` (`signal_id`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `svsignal_ivalue`
--
ALTER TABLE `svsignal_ivalue`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `svsignal_ivalue`
--
ALTER TABLE `svsignal_ivalue`
  ADD CONSTRAINT `svsignal_ivalue_ibfk_1` FOREIGN KEY (`signal_id`) REFERENCES `svsignal_signal` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;


-- phpMyAdmin SQL Dump
-- version 5.1.1
-- https://www.phpmyadmin.net/
--
-- Хост: mysql
-- Время создания: Дек 15 2021 г., 04:31
-- Версия сервера: 10.5.8-MariaDB-1:10.5.8+maria~focal-log
-- Версия PHP: 7.4.20

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- База данных: `insiteexpert_v4`
--

-- --------------------------------------------------------

--
-- Структура таблицы `svsignal_fvalue`
--

CREATE TABLE `svsignal_fvalue` (
  `id` int(11) NOT NULL,
  `signal_id` int(11) NOT NULL,
  `utime` int(11) NOT NULL,
  `value` double NOT NULL,
  `offline` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `svsignal_fvalue`
--
ALTER TABLE `svsignal_fvalue`
  ADD PRIMARY KEY (`id`),
  ADD KEY `signal_id` (`signal_id`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `svsignal_fvalue`
--
ALTER TABLE `svsignal_fvalue`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `svsignal_fvalue`
--
ALTER TABLE `svsignal_fvalue`
  ADD CONSTRAINT `svsignal_fvalue_ibfk_1` FOREIGN KEY (`signal_id`) REFERENCES `svsignal_signal` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;

-- phpMyAdmin SQL Dump
-- version 5.1.1
-- https://www.phpmyadmin.net/
--
-- Хост: mysql
-- Время создания: Дек 15 2021 г., 04:31
-- Версия сервера: 10.5.8-MariaDB-1:10.5.8+maria~focal-log
-- Версия PHP: 7.4.20

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- База данных: `insiteexpert_v4`
--

-- --------------------------------------------------------

--
-- Структура таблицы `svsignal_mvalue`
--

CREATE TABLE `svsignal_mvalue` (
  `id` int(11) NOT NULL,
  `signal_id` int(11) NOT NULL,
  `utime` int(11) NOT NULL,
  `max` double NOT NULL,
  `min` double NOT NULL,
  `mean` double NOT NULL,
  `median` double NOT NULL,
  `offline` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `svsignal_mvalue`
--
ALTER TABLE `svsignal_mvalue`
  ADD KEY `signal_id` (`signal_id`);

--
-- Ограничения внешнего ключа сохраненных таблиц
--

--
-- Ограничения внешнего ключа таблицы `svsignal_mvalue`
--
ALTER TABLE `svsignal_mvalue`
  ADD CONSTRAINT `svsignal_mvalue_ibfk_1` FOREIGN KEY (`signal_id`) REFERENCES `svsignal_signal` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;