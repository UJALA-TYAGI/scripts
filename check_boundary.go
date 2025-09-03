func (tokenSuite *TokenTestSuite) TestValidateRolesForAudience() {
    // Setup: inject some audience-role mappings
    environ.EnvInstance.AudienceRoles = map[string][]string{
        "aud1": {"ROLE_A", "ROLE_B"},
        "aud2": {"ROLE_X"},
    }

    tokenSuite.Run("Valid audience and allowed roles", func() {
        err := token.ValidateRolesForAudience("aud1", []string{"ROLE_A"})
        tokenSuite.NoError(err)
    })

    tokenSuite.Run("Invalid audience", func() {
        err := token.ValidateRolesForAudience("unknown_aud", []string{"ROLE_A"})
        var customErr *customError.InternalError
        tokenSuite.Error(err)
        tokenSuite.True(errors.As(err, &customErr))
        tokenSuite.Equal(customError.InvalidAudience, customErr.Code)
    })

    tokenSuite.Run("Valid audience but disallowed role", func() {
        err := token.ValidateRolesForAudience("aud1", []string{"ROLE_X"})
        var customErr *customError.InternalError
        tokenSuite.Error(err)
        tokenSuite.True(errors.As(err, &customErr))
        tokenSuite.Equal(customError.RoleAudienceMismatch, customErr.Code)
    })

    tokenSuite.Run("Multiple roles, some invalid", func() {
        err := token.ValidateRolesForAudience("aud1", []string{"ROLE_A", "ROLE_X"})
        var customErr *customError.InternalError
        tokenSuite.Error(err)
        tokenSuite.True(errors.As(err, &customErr))
        tokenSuite.Equal(customError.RoleAudienceMismatch, customErr.Code)
    })

    tokenSuite.Run("Multiple roles, all valid", func() {
        err := token.ValidateRolesForAudience("aud1", []string{"ROLE_A", "ROLE_B"})
        tokenSuite.NoError(err)
    })
}
