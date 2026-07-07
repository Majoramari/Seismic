package icu.seismic.plugin.settings

import com.intellij.openapi.options.Configurable
import com.intellij.util.ui.FormBuilder
import icu.seismic.plugin.SeismicSettings
import javax.swing.JCheckBox
import javax.swing.JComponent
import javax.swing.JPasswordField
import javax.swing.JTextField

class SeismicConfigurable : Configurable {
    private val apiKeyField = JPasswordField()
    private val apiUrlField = JTextField()
    private val enabledCheckbox = JCheckBox("Enable tracking")
    private val statusBarCheckbox = JCheckBox("Show status bar widget")

    private var panel: JComponent? = null

    override fun getDisplayName(): String = "Seismic"

    override fun createComponent(): JComponent {
        apiKeyField.text = SeismicSettings.getApiKey()
        apiUrlField.text = SeismicSettings.getApiUrl()
        enabledCheckbox.isSelected = SeismicSettings.isEnabled()
        statusBarCheckbox.isSelected = SeismicSettings.isStatusBarEnabled()

        val built = FormBuilder.createFormBuilder()
            .addLabeledComponent("API Key:", apiKeyField)
            .addLabeledComponent("API URL:", apiUrlField)
            .addComponent(enabledCheckbox)
            .addComponent(statusBarCheckbox)
            .panel

        panel = built
        return built
    }

    override fun isModified(): Boolean {
        return String(apiKeyField.password) != SeismicSettings.getApiKey() ||
                apiUrlField.text != SeismicSettings.getApiUrl() ||
                enabledCheckbox.isSelected != SeismicSettings.isEnabled() ||
                statusBarCheckbox.isSelected != SeismicSettings.isStatusBarEnabled()
    }

    override fun apply() {
        SeismicSettings.setApiKey(String(apiKeyField.password))
        SeismicSettings.setApiUrl(apiUrlField.text)
        SeismicSettings.setEnabled(enabledCheckbox.isSelected)
        SeismicSettings.setStatusBarEnabled(statusBarCheckbox.isSelected)
    }

    override fun reset() {
        apiKeyField.text = SeismicSettings.getApiKey()
        apiUrlField.text = SeismicSettings.getApiUrl()
        enabledCheckbox.isSelected = SeismicSettings.isEnabled()
        statusBarCheckbox.isSelected = SeismicSettings.isStatusBarEnabled()
    }
}