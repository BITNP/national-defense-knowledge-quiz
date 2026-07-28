#!/usr/bin/env python3
"""Convert old problems.sql INSERT statements to JSON config format."""
import re
import json
import sys

def parse_sql_inserts(sql_path):
    problems = []
    with open(sql_path, 'r', encoding='utf-8') as f:
        content = f.read()

    pattern = r"\(\s*(\d+)\s*,\s*'([^']*)'\s*,\s*'((?:[^']|'')*)'\s*,\s*'((?:[^']|'')*)'\s*,\s*'((?:[^']|'')*)'\s*,\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)"
    for m in re.finditer(pattern, content):
        pid, ptype, text, data_raw, answer, score, exam, active = m.groups()
        text = text.replace("''", "'")
        data_raw = data_raw.replace("''", "'")
        answer = answer.replace("''", "'")
        try:
            data = json.loads(data_raw)
        except json.JSONDecodeError:
            data = data_raw.split(",")
        problems.append({
            "type": ptype,
            "text": text,
            "data": data,
            "answer": answer,
            "score": int(score),
            "active": bool(int(active)),
        })
    return problems

def make_exam_json(exam_sql_path, prize_sql_path, problems_sql_path, output_path):
    # Read exam data from exam-2025.sql
    exam = {
        "title": "2025北京理工大学国家网络安全宣传周线上答题",
        "intro": "北京理工大学2025国家网络安全宣传周线上答题，由网络开拓者协会承办。开始答题后将持续计时，若意外退出，可重新进入继续答题。每人有两次答题机会，参加周五网安俱乐部开展的网络安全知识讲座可获得额外两次答题机会。 <br/>请认真填写学号姓名，每次答题提交后即自动抽奖，中奖概率与分数正相关。奖品请于2025年9月22日 - 9月28日 19:00-22:00到网络开拓者协会办公室(北校区北理桥下左数第一个)领取哦~ <br/>一等奖：神秘大奖 1份<br/>二等奖：闪迪64GB U盘 10份<br/>三等奖：USB扩展坞 10份<br/>四等奖：手提电脑包 15份<br/>欢乐奖：树莓娘通行证 110份<br/>网安俱乐部外场活动时空坐标: 2025-09-19（周五）19:00-21:00, 地点: 综A102<br/>PS：网安周线上答题活动持续时间为2025.9.15-9.21。网安外场活动中也可线下参与答题抽奖活动（包括额外抽奖次数）",
        "limit_time": 300,
        "random": 15,
        "limit_number": 2,
        "active": True
    }

    prizes = [
        {"text": "神秘大奖", "remain": 1},
        {"text": "闪迪64GB U盘", "remain": 10},
        {"text": "USB扩展坞", "remain": 10},
        {"text": "手提电脑包", "remain": 15},
        {"text": "树莓娘通行证", "remain": 110},
    ]

    problems = parse_sql_inserts(problems_sql_path)

    data = {
        "exam": exam,
        "problems": problems,
        "prizes": prizes,
    }

    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=2)

    print(f"Written {len(problems)} problems to {output_path}")

if __name__ == "__main__":
    import os
    script_dir = os.path.dirname(os.path.abspath(__file__))
    project_dir = os.path.dirname(script_dir)
    ndquiz_dir = os.path.dirname(project_dir)
    make_exam_json(
        exam_sql_path=os.path.join(ndquiz_dir, "OnlineExam", "db", "exam-2025.sql"),
        prize_sql_path=os.path.join(ndquiz_dir, "OnlineExam", "db", "prize-2025.sql"),
        problems_sql_path=os.path.join(ndquiz_dir, "OnlineExam", "db", "problems.sql"),
        output_path=os.path.join(project_dir, "config", "exam.json"),
    )
