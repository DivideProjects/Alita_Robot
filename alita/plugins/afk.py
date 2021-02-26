# Copyright (C) 2020 - 2021 Divkix. All rights reserved. Source code available under the AGPL.
#
# This file is part of Alita_Robot.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.

# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.


from pyrogram import filters
from pyrogram.types import Message

from alita import LOGGER, PREFIX_HANDLER
from alita.bot_class import Alita
from alita.database.afk_db import AFK
from alita.tr_engine import tlang

__PLUGIN__ = "AFK"

__help__ = """
Module for enabling auto replies when you are AFK.
When enabled,
anyone who mentions you will be replied with a message saying that
you are AFK.

**Setting AFK Status**
 × /afk <reason>
Enable auto replies when you are AFK.
To stop it, send message to any group.

* Reason is optional
"""

# Initialise
db = AFK()

@Alita.on_message(
    filters.command("afk", PREFIX_HANDLER) & filters.group,
)
async def set_afk(_, m: Message):

    afkmsg = f"User {(await mention_html(m.from_user.first_name, m.from_user.id))} is now afk!"

    if len(m.text.split()) > 1:
        reason = "\n<b>Reason:</b>" + m.text.split(None, 1)[1]
    else:
        reason = ""

    try:
        await db.add_afk(m.from_user.id, reason)
        replymsg = await m.reply_text(afkmsg+reason)
    except Exception as ef:
        await m.reply_text(ef)
        LOGGER.error(ef)

    return


@Alita.on_message(filters.mentioned & filters.group & ~filters.bot, group=11)
async def afk_mentioned(c: Alita, m: Message):
    if m.from_user:
        try:
            user_afk = await db.check_afk(m.from_user.id)
        except Exception as ef:
            await m.reply_text(f"Error while chekcing afk\n{ef}")
            return
 
        if not user_afk:
            return

        afkmsg = f"{(await c.get_users(user_afk['user_id'])).first_name} is Afk!"

        if user_afk['reason']:
            afkmsg += f"<b>Reason:</b> {user_afk['reason']}"

        await m.reply_text(afkmsg)
    return


@Alita.on_message(filters.group, group=12)
async def rem_afk: Alita, m: Message):
    if m.from_user:
        try:
            user_afk = await db.check_afk(m.from_user.id)
        except Exception as ef:
            await m.reply_text(f"Error while chekcing afk\n{ef}")
            return

        if user_afk:
            await db.remove_afk(m.from_user.id)
            await m.reply_text(f"{(await c.get_users(user_afk['user_id'])).first_name} is no longer Afk!")

    return
