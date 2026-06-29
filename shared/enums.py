from enum import StrEnum


class Gender(StrEnum):
    MALE = "male"
    FEMALE = "female"


class Language(StrEnum):
    RU = "ru"
    EN = "en"


class NsfwLevel(StrEnum):
    SAFE = "safe"
    ADULT = "adult"
