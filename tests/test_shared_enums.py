from shared.enums import Gender, Language, NsfwLevel


def test_enum_values() -> None:
    assert Gender.MALE == "male"
    assert Language.EN == "en"
    assert NsfwLevel.SAFE == "safe"
