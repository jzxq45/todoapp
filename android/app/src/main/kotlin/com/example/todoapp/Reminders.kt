package com.example.todoapp

import android.app.AlarmManager
import android.app.PendingIntent
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.SharedPreferences
import android.net.Uri
import androidx.core.app.NotificationManagerCompat
import org.json.JSONArray
import org.json.JSONObject

class ReminderReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        ReminderScheduler.process(context)
    }
}

class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        ReminderScheduler.process(context)
    }
}

private data class Reminder(
    val id: String,
    val title: String,
    val message: String,
    val triggerAt: Long,
    val sent: Boolean = false,
    val attempts: Int = 0
)

object ReminderScheduler {
    private const val PREFS = "reminder_state"
    private const val MAX_ATTEMPTS = 10
    private const val RETRY_MS = 5 * 60_000L

    /** Полная синхронизация напоминаний из приложения: заменяет состояние и обрабатывает его. */
    fun sync(context: Context, payload: String) {
        val incoming = try {
            parse(payload)
        } catch (_: Exception) {
            return
        }
        val prefs = prefs(context)
        val existing = load(prefs)
        val oldByID = existing.associateBy { it.id }
        val merged = incoming.map { reminder ->
            val old = oldByID[reminder.id]
            if (old != null && old.triggerAt == reminder.triggerAt && old.sent) {
                reminder.copy(sent = true, attempts = old.attempts)
            } else {
                reminder
            }
        }
        val removed = existing.filter { old -> incoming.none { it.id == old.id } }
        save(prefs, merged)
        removed.forEach { cancelAlarm(context, it.id) }
        process(context)
    }

    /** Проверка доставки: отправляет всё неотправленное с наступившим сроком и перепоставляет будущие будильники. */
    fun process(context: Context) {
        NotificationHelper.createNotificationChannel(context)
        val prefs = prefs(context)
        val now = System.currentTimeMillis()
        for (reminder in load(prefs)) {
            when {
                reminder.sent -> Unit
                reminder.triggerAt <= now -> deliver(context, prefs, reminder)
                else -> setAlarm(context, reminder.id, reminder.triggerAt)
            }
        }
    }

    private fun deliver(context: Context, prefs: SharedPreferences, reminder: Reminder) {
        if (NotificationManagerCompat.from(context).areNotificationsEnabled()) {
            val intent = Intent(context, MainActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
            }
            NotificationHelper.showNotification(
                context,
                reminder.id.hashCode() and 0x7fffffff,
                reminder.title,
                reminder.message,
                intent
            )
            save(prefs, load(prefs).map { if (it.id == reminder.id) it.copy(sent = true) else it })
            cancelAlarm(context, reminder.id)
            return
        }
        val updated = reminder.copy(attempts = reminder.attempts + 1)
        save(prefs, load(prefs).map { if (it.id == updated.id) updated else it })
        if (updated.attempts < MAX_ATTEMPTS) {
            setAlarm(context, updated.id, System.currentTimeMillis() + RETRY_MS)
        }
    }

    private fun setAlarm(context: Context, id: String, at: Long) {
        val alarm = context.getSystemService(Context.ALARM_SERVICE) as AlarmManager
        try {
            alarm.setExactAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, at, pendingIntent(context, id))
        } catch (_: SecurityException) {
            alarm.setAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, at, pendingIntent(context, id))
        }
    }

    private fun pendingIntent(context: Context, id: String): PendingIntent {
        val intent = Intent(context, ReminderReceiver::class.java).apply {
            data = Uri.parse("todoapp://reminder/$id")
        }
        return PendingIntent.getBroadcast(
            context,
            id.hashCode(),
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    private fun cancelAlarm(context: Context, id: String) {
        val alarm = context.getSystemService(Context.ALARM_SERVICE) as AlarmManager
        alarm.cancel(pendingIntent(context, id))
    }

    private fun prefs(context: Context): SharedPreferences =
        context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)

    private fun parse(payload: String): List<Reminder> {
        val array = JSONArray(payload)
        val reminders = ArrayList<Reminder>(array.length())
        for (index in 0 until array.length()) {
            val item = array.getJSONObject(index) ?: continue
            val id = item.optString("id")
            if (id.isEmpty()) continue
            reminders.add(
                Reminder(
                    id = id,
                    title = item.optString("title"),
                    message = item.optString("message"),
                    triggerAt = item.optLong("triggerAt")
                )
            )
        }
        return reminders
    }

    private fun load(prefs: SharedPreferences): List<Reminder> {
        val raw = prefs.getString(KEY, null) ?: return emptyList()
        return try {
            val array = JSONArray(raw)
            val reminders = ArrayList<Reminder>(array.length())
            for (index in 0 until array.length()) {
                val item = array.getJSONObject(index)
                reminders.add(
                    Reminder(
                        id = item.getString("id"),
                        title = item.getString("title"),
                        message = item.getString("message"),
                        triggerAt = item.getLong("triggerAt"),
                        sent = item.optBoolean("sent"),
                        attempts = item.optInt("attempts")
                    )
                )
            }
            reminders
        } catch (_: Exception) {
            emptyList()
        }
    }

    private fun save(prefs: SharedPreferences, reminders: List<Reminder>) {
        val array = JSONArray()
        reminders.forEach { reminder ->
            array.put(
                JSONObject()
                    .put("id", reminder.id)
                    .put("title", reminder.title)
                    .put("message", reminder.message)
                    .put("triggerAt", reminder.triggerAt)
                    .put("sent", reminder.sent)
                    .put("attempts", reminder.attempts)
            )
        }
        prefs.edit().putString(KEY, array.toString()).apply()
    }

    private const val KEY = "items"
}
