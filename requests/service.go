package requests

import "mime/multipart"

type CreateServiceRequest struct {
    // Basic Information
    Name       string                `form:"name" binding:"required"`
    Slug       string                `form:"slug"`
    Logo       *multipart.FileHeader `form:"logo"`
    Icon       *multipart.FileHeader `form:"icon"`
    IconPublicID *string             `form:"icon_public_id`
    LogoPublicID *string             `form:"logo_public_id`
    CategoryID uint                  `form:"category_id" binding:"required"`

    // Description
    Description *string `form:"description"`
    HowToTopup  *string `form:"how_to_topup"`
    Notes       *string `form:"notes"`

    // Customer Number Format
    CustomerNoFormat string `form:"customer_no_format" binding:"required"`

    // Format Configuration
    ExampleFormat     *string `form:"example_format"`
    Field1Label       string  `form:"field1_label" binding:"required"`
    Field1Placeholder string  `form:"field1_placeholder" binding:"required"`
    Field2Label       *string `form:"field2_label"`
    Field2Placeholder *string `form:"field2_placeholder"`

    // Flags
    IsActive  bool `form:"is_active"`
    IsPopular bool `form:"is_popular"`
}

type UpdateServiceRequest struct {
    // Basic Information
    Name       string                `form:"name"`
    Slug       string                `form:"slug"`
    Logo       *multipart.FileHeader `form:"logo"`
    Icon       *multipart.FileHeader `form:"icon"`
    RemoveLogo bool                  `form:"remove_logo"`
    RemoveIcon bool                  `form:"remove_icon"`
    CategoryID uint                  `form:"category_id"`

    // Description
    Description *string `form:"description"`
    HowToTopup  *string `form:"how_to_topup"`
    Notes       *string `form:"notes"`

    // Customer Number Format
    CustomerNoFormat string `form:"customer_no_format"`

    // Format Configuration
    ExampleFormat     *string `form:"example_format"`
    Field1Label       string  `form:"field1_label"`
    Field1Placeholder string  `form:"field1_placeholder"`
    Field2Label       *string `form:"field2_label"`
    Field2Placeholder *string `form:"field2_placeholder"`

    // Flags
    IsActive  *bool `form:"is_active"`
    IsPopular *bool `form:"is_popular"`
}